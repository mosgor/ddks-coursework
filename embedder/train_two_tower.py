#!/usr/bin/env python3
"""
Two-Tower Recommendation Model Training Script
- Item Tower: SentenceTransformer (text -> embedding)
- User Tower: MLP (behavior history -> embedding)
- Loss: TripletMarginLoss (contrastive learning)
- Export: ONNX for Go inference
"""
import os
import json
import logging
import numpy as np
import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.utils.data import Dataset, DataLoader
from sentence_transformers import SentenceTransformer
import psycopg2
from tqdm import tqdm
import onnx

# ─── 12-FACTOR: Configuration via Environment Variables ───────────────────────
DB_URL = os.getenv("DATABASE_URL")
if not DB_URL:
    raise ValueError("DATABASE_URL environment variable is required")

ITEM_MODEL_NAME = os.getenv("ITEM_MODEL", "all-MiniLM-L6-v2")
EMBEDDING_DIM = int(os.getenv("EMBEDDING_DIM", "384"))
BATCH_SIZE = int(os.getenv("BATCH_SIZE", "64"))
EPOCHS = int(os.getenv("EPOCHS", "5"))
LEARNING_RATE = float(os.getenv("LEARNING_RATE", "1e-3"))
MARGIN = float(os.getenv("TRIPLET_MARGIN", "0.5"))
OUTPUT_DIR = os.getenv("OUTPUT_DIR", "./models")
os.makedirs(OUTPUT_DIR, exist_ok=True)

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
logger = logging.getLogger(__name__)

# ─── DATA LOADING ─────────────────────────────────────────────────────────────
def load_interactions(conn):
    """Load user interactions and event texts from PostgreSQL."""
    with conn.cursor() as cur:
        cur.execute("""
            SELECT ub.user_id, ub.event_id, ub.event_type, e.title, 
                   COALESCE(e.description, ''), COALESCE(e.genre, ''), 
                   COALESCE(e.tags, '[]')::jsonb
            FROM user_behavior ub
            JOIN events e ON ub.event_id = e.id
            WHERE e.embedding IS NOT NULL
        """)
        rows = cur.fetchall()
    
    # Group by user
    user_histories = {}
    for uid, eid, etype, title, desc, genre, tags in rows:
        if uid not in user_histories:
            user_histories[uid] = {"events": [], "weights": [], "texts": []}
        
        # Weight by interaction type
        weight = {"view": 1.0, "cart_add": 2.0, "purchase": 3.0}.get(etype, 1.0)
        user_histories[uid]["events"].append(eid)
        user_histories[uid]["weights"].append(weight)
        
        tags_list = ", ".join(tags) if isinstance(tags, list) else ""
        text = f"{title}. {desc}. Жанр: {genre}. Теги: {tags_list}"
        user_histories[uid]["texts"].append(text)
        
    return user_histories

def load_all_events(conn):
    """Load all events for negative sampling."""
    with conn.cursor() as cur:
        cur.execute("SELECT id, title, COALESCE(description,''), COALESCE(genre,''), COALESCE(tags,'[]')::jsonb FROM events WHERE embedding IS NOT NULL")
        return [(row[0], row[1], row[2], row[3], row[4]) for row in cur.fetchall()]

# ─── MODELS ───────────────────────────────────────────────────────────────────
class ItemTower(nn.Module):
    """Freezes pre-trained SentenceTransformer. Used only for inference."""
    def __init__(self, model_name):
        super().__init__()
        self.model = SentenceTransformer(model_name, device="cpu")
        # Disable gradient updates for Item Tower
        for param in self.model.parameters():
            param.requires_grad = False

    @torch.no_grad()
    def forward(self, texts):
        return self.model.encode(texts, convert_to_tensor=True)

class UserTower(nn.Module):
    """Maps aggregated interaction features to embedding space."""
    def __init__(self, input_dim, hidden_dim, output_dim):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(input_dim, hidden_dim),
            nn.LayerNorm(hidden_dim),
            nn.GELU(),
            nn.Dropout(0.2),
            nn.Linear(hidden_dim, output_dim),
            nn.LayerNorm(output_dim)
        )

    def forward(self, x):
        return F.normalize(self.net(x), p=2, dim=1)

# ─── DATASET & DATALOADER ─────────────────────────────────────────────────────
class TripletDataset(Dataset):
    def __init__(self, user_histories, item_embeddings, event_ids, neg_sample_size=3):
        self.user_histories = user_histories
        self.item_emb = item_embeddings  # dict: event_id -> tensor
        self.event_ids = list(event_ids)
        self.neg_sample_size = neg_sample_size

    def __len__(self):
        return len(self.user_histories)

    def __getitem__(self, idx):
        uid = list(self.user_histories.keys())[idx]
        data = self.user_histories[uid]
        pos_embs = torch.stack([self.item_emb[eid] for eid in data["events"]])
        weights = torch.tensor(data["weights"], dtype=torch.float32).unsqueeze(1)
        
        # Anchor: weighted average of interacted item embeddings
        anchor = (pos_embs * weights).sum(dim=0) / weights.sum()
        anchor = F.normalize(anchor, p=2, dim=0)
        
        # Positive: random interacted item
        pos_idx = np.random.randint(len(data["events"]))
        positive = F.normalize(self.item_emb[data["events"][pos_idx]], p=2, dim=0)
        
        # Negatives: random items NOT in history
        negs = []
        for _ in range(self.neg_sample_size):
            while True:
                neg_id = np.random.choice(self.event_ids)
                if neg_id not in data["events"]:
                    negs.append(F.normalize(self.item_emb[neg_id], p=2, dim=0))
                    break
        negatives = torch.stack(negs)
        
        return anchor, positive, negatives

# ─── TRAINING LOOP ────────────────────────────────────────────────────────────
def train(user_tower, loader, optimizer, device):
    criterion = nn.TripletMarginLoss(margin=MARGIN)
    user_tower.train()
    total_loss = 0.0
    num_batches = 0
    
    for anchors, positives, negatives in tqdm(loader, desc="Training"):
        anchors, positives, negatives = (
            anchors.to(device), positives.to(device), negatives.to(device)
        )
        
        optimizer.zero_grad()
        user_vecs = user_tower(anchors)
        
        # Compute loss against multiple negatives
        loss = 0.0
        for i in range(negatives.shape[1]):
            loss += criterion(user_vecs, positives, negatives[:, i, :])
        loss /= negatives.shape[1]
        
        loss.backward()
        optimizer.step()
        total_loss += loss.item()
        num_batches += 1

    if num_batches == 0:
        logger.warning("No batches processed in training loop")
        return 0.0
        
    return total_loss / len(loader)

# ─── ONNX EXPORT ──────────────────────────────────────────────────────────────
def export_user_tower(model, output_path, dummy_input):
    model.eval()
    torch.onnx.export(
        model,
        dummy_input,
        output_path,
        export_params=True,
        opset_version=14,
        do_constant_folding=True,
        input_names=["anchor"],
        output_names=["user_embedding"],
        dynamic_axes={"anchor": {0: "batch_size"}, "user_embedding": {0: "batch_size"}}
    )
    logger.info(f"User Tower exported to {output_path}")

# ─── DB UPDATE ────────────────────────────────────────────────────────────────
def update_event_embeddings(conn, item_tower, events_data):
    """Compute fresh item embeddings and store in PG."""
    logger.info("Updating event embeddings...")
    with conn.cursor() as cur:
        for eid, title, desc, genre, tags in tqdm(events_data, desc="Encoding events"):
            tags_str = ", ".join(tags) if isinstance(tags, list) else ""
            text = f"{title}. {desc}. Жанр: {genre}. Теги: {tags_str}"
            emb = item_tower.forward([text]).cpu().numpy()[0]
            emb = emb / np.linalg.norm(emb)
            cur.execute("UPDATE events SET embedding = %s WHERE id = %s", (emb.tolist(), eid))
    conn.commit()
    logger.info("Event embeddings updated.")

# ─── MAIN ─────────────────────────────────────────────────────────────────────
def main():
    logger.info("🚀 Starting Two-Tower Training Pipeline")
    
    # 1. Connect to DB
    conn = psycopg2.connect(DB_URL)
    logger.info("📊 Loading data from PostgreSQL...")
    user_histories = load_interactions(conn)
    events_data = load_all_events(conn)
    
    if not user_histories:
        logger.warning("No interactions found. Skipping training.")
        conn.close()
        return

    # 2. Initialize Item Tower (CPU-only for inference)
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    logger.info(f"📦 Loading Item Tower: {ITEM_MODEL_NAME} on {device}")
    item_tower = ItemTower(ITEM_MODEL_NAME)
    
    # Precompute item embeddings
    all_ids = [e[0] for e in events_data]
    texts = [f"{e[1]}. {e[2]}. Жанр: {e[3]}. Теги: {', '.join(e[4]) if isinstance(e[4], list) else ''}" for e in events_data]
    item_embs = {eid: emb for eid, emb in zip(all_ids, item_tower.forward(texts).cpu())}
    
    # 3. Initialize User Tower & DataLoader
    dataset = TripletDataset(user_histories, item_embs, all_ids, neg_sample_size=3)
    loader = DataLoader(dataset, batch_size=BATCH_SIZE, shuffle=True)
    user_tower = UserTower(input_dim=EMBEDDING_DIM, hidden_dim=256, output_dim=EMBEDDING_DIM).to(device)
    optimizer = torch.optim.AdamW(user_tower.parameters(), lr=LEARNING_RATE, weight_decay=1e-4)
    
    # 4. Train
    for epoch in range(1, EPOCHS + 1):
        loss = train(user_tower, loader, optimizer, device)
        logger.info(f"Epoch {epoch}/{EPOCHS} - Loss: {loss:.4f}")
    
    # 5. Export User Tower to ONNX
    onnx_path = os.path.join(OUTPUT_DIR, "user_tower.onnx")
    dummy_input = torch.randn(1, EMBEDDING_DIM)
    export_user_tower(user_tower, onnx_path, dummy_input)
    
    # 6. Update DB
    update_event_embeddings(conn, item_tower, events_data)
    
    conn.close()
    logger.info("✅ Training & Export completed successfully.")

if __name__ == "__main__":
    main()