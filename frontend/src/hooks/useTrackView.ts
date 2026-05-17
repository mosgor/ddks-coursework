import { useEffect } from 'react';
import { api } from '../utils/api';

export const useTrackView = (eventId: number, token?: string) => {
  useEffect(() => {
    if (token && eventId) {
      api.track(eventId, 'view', token).catch(console.error);
    }
  }, [eventId, token]);
};