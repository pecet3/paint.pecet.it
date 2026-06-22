import { BrowserRouter, Route, Routes } from 'react-router'
import { Home } from './pages/Home'
import Navbar from './components/Navbar'

function AppContent() {


    return (
        <>
            <Navbar />
            <Routes>
                <Route path="/" element={
                    <>
                        <Home />
                        <div className='h-32' />
                    </>
                } />
            </Routes>


        </>
    )
}

function App() {
    return (
        <BrowserRouter>
            <AppContent />
        </BrowserRouter>
    )
}

export default App