import { BrowserRouter, Route, Routes, useNavigate, useLocation } from 'react-router'
import { Home } from './pages/Home'
import Navbar from './components/Navbar'
import { useStore } from './Store'
import { useEffect } from 'react'
import type { PhotoRecord } from './pocketbase'
import { PhotoModal } from './components/PhotoModal'

function AppContent() {
    const {
        photoModal,
        setPhotoModal,
        photosFeed,
        fetchSinglePhoto
    } = useStore()

    const navigate = useNavigate()
    const location = useLocation()
    const searchParams = new URLSearchParams(location.search)
    const photoIdFromUrl = searchParams.get('photoId')

    useEffect(() => {
        if (location.pathname === '/photos') {
            if (photoModal && photoIdFromUrl !== photoModal.id) {
                navigate(`/photos?photoId=${photoModal.id}`, { replace: true })
            } else if (!photoModal && photoIdFromUrl) {
                navigate('/photos', { replace: true })
            }
        }
    }, [photoModal, location.pathname])

    useEffect(() => {
        if (location.pathname === '/photos' && photoIdFromUrl) {
            const existingPhoto = photosFeed.find(
                (item) => item.id === photoIdFromUrl && !('photos' in item)
            )

            if (existingPhoto) {
                setPhotoModal(existingPhoto as PhotoRecord)
            } else {
                fetchSinglePhoto(photoIdFromUrl).then((fetchedPhoto) => {
                    if (fetchedPhoto) {
                        setPhotoModal(fetchedPhoto)
                    } else {
                        navigate('/photos', { replace: true })
                    }
                })
            }
        } else if (location.pathname === '/photos' && !photoIdFromUrl && photoModal) {
            setPhotoModal(null)
        }
    }, [photoIdFromUrl, location.pathname])

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

            {photoModal && (
                <PhotoModal
                    onClose={() => setPhotoModal(null)}
                    photo={photoModal}
                />
            )}
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