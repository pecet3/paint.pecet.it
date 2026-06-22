import React, { createContext, useContext, useState, type ReactNode } from 'react';





interface StoreContextType {
    loading: boolean;
    error: string | null;
}

const StoreContext = createContext<StoreContextType | undefined>(undefined);

interface StoreProviderProps {
    children: ReactNode;
}


export const StoreProvider: React.FC<StoreProviderProps> = ({ children }) => {
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);


    return (
        <StoreContext.Provider value={{
            loading,
            error,

        }}>
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