import { BrowserRouter, Navigate, Route, Routes } from 'react-router'
import { Home } from './pages/Home'
import Navbar from './components/Navbar'
import { StoreProvider, useStore } from './Store'
import { Login, LoginForm } from './pages/Login';


interface ProtectedRouteProps {
    children: React.ReactElement;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
    const { user, loading } = useStore();

    if (loading) {
        return (
            <div className="flex items-center justify-center h-screen font-sans">
                <p className="text-gray-500">Loading application state...</p>
            </div>
        );
    }

    if (!user) {
        return <Navigate to="/login" replace />;
    }

    return children;
};

function AppContent() {


    return (
        <>
            <Navbar />
            <Routes>
                <Route path="/login" element={<Login />} />
                <Route path="/" element={
                    <ProtectedRoute>
                        <>
                            <Home />
                            <div className='h-32' />
                        </>
                    </ProtectedRoute>
                } />
            </Routes>


        </>
    )
}

function App() {
    return (
        <StoreProvider>
            <BrowserRouter>
                <AppContent />
            </BrowserRouter>
        </StoreProvider>
    )
}

export default App