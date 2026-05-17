import React, { useEffect, useState } from 'react';
import { api } from '../../utils/api';
import { useStateContext } from '../../contexts';
import { EventCard } from '../EventCard';
import styles from '../../styles/home.module.css';

export const RecommendationSection: React.FC = () => {
  const { state } = useStateContext();
  const [events, setEvents] = useState<any[]>([]);

  useEffect(() => {
    if (!state.token) return;
    api.getRecommendations(state.token)
      .then(res => setEvents(res.events))
      .catch(console.error);
  }, [state.token]);

  if (!events.length) return null;

  return (
    <div className={styles.home__section}>
      <h2 className={styles.home__sectionTitle}>Рекомендуем вам</h2>
      <div className={styles.home__list}>
        {events.map(ev => (
          <EventCard key={ev.id} event={ev} />
        ))}
      </div>
    </div>
  );
};