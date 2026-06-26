import React, { createContext, useContext, useState, useEffect, type ReactNode } from 'react';

interface User {
    uuid: string;
    name: string;
}

interface StoreContextType {
    user: User | null;
    loading: boolean;
    error: string | null;
    setUser: (user: User | null) => void;
    checkAuth: () => Promise<void>;
}

const StoreContext = createContext<StoreContextType | undefined>(undefined);

interface StoreProviderProps {
    children: ReactNode;
}

export const StoreProvider: React.FC<StoreProviderProps> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    const checkAuth = async () => {
        try {
            setLoading(true);
            const response = await fetch('/ping', { credentials: 'include' });

            if (response.ok) {
                const data: User = await response.json();
                setUser(data);
            } else {
                setUser(null);
            }
        } catch (err) {
            setError('Failed to authenticate connection');
            setUser(null);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        checkAuth();
    }, []);

    return (
        <StoreContext.Provider value={{ user, loading, error, setUser, checkAuth }}>
            {children}
        </StoreContext.Provider>
    );
};

export const useStore = () => {
    const context = useContext(StoreContext);
    if (context === undefined) {
        throw new Error('useStore must be used within a StoreProvider');
    }
    return context;
};