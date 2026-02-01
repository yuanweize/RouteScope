/* eslint-disable react-refresh/only-export-components */
import React, { createContext, useContext, useMemo, useState, useCallback } from 'react';
import { getTargets } from '../api';
import type { Target } from '../api';

interface AppContextValue {
  isDark: boolean;
  setIsDark: (value: boolean) => void;
  targets: Target[];
  selectedTarget: string;
  setSelectedTarget: (value: string) => void;
  refreshTargets: () => Promise<void>;
}

const AppContext = createContext<AppContextValue | undefined>(undefined);

export const AppContextProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isDark, setIsDark] = useState(false);
  const [targets, setTargets] = useState<Target[]>([]);
  const [selectedTarget, setSelectedTarget] = useState('');

  const refreshTargets = useCallback(async () => {
    try {
      const data = (await getTargets()) as Target[];
      setTargets(data);
      if (data.length > 0 && !selectedTarget) {
        setSelectedTarget(data[0].address || '');
      }
    } catch (e) {
      console.error(e);
    }
  }, [selectedTarget]);

  const value = useMemo(
    () => ({ isDark, setIsDark, targets, selectedTarget, setSelectedTarget, refreshTargets }),
    [isDark, targets, selectedTarget, refreshTargets]
  );

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};

export const useAppContext = () => {
  const ctx = useContext(AppContext);
  if (!ctx) {
    throw new Error('useAppContext must be used within AppContextProvider');
  }
  return ctx;
};
