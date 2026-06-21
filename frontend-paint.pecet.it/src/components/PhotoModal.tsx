import { pb } from "../pocketbase"
import type { Photo } from "../Store"

interface ImageModalProps {
    url?: string
    photo: Photo
    onClose: () => void
}

export const PhotoModal = ({ url, photo, onClose }: ImageModalProps) => {
    const imageUrl = pb.files.getURL(photo, photo.image)

    const formattedDate = photo.date
        ? new Date(photo.date).toLocaleDateString('pl-PL', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        })
        : null

    return (
        <div
            onClick={onClose}
            className="fixed inset-0 z-50 flex items-center justify-center p-4
             bg-black/80 backdrop-blur-md transition-all duration-300 animate-fade-in"
        >
            <div
                onClick={(e) => e.stopPropagation()}
                className="relative flex flex-col md:flex-row w-auto
                max-w-6xl max-h-[90vh] md:max-h-[85vh] overflow-y-auto 
                md:overflow-hidden rounded-xl bg-neutral-950 border border-neutral-800 shadow-2xl"
            >
                <button
                    onClick={onClose}
                    className="absolute -top-1 -right-1 z-50 p-2 rounded-full bg-black/50 hover:cursor-pointer
                     text-neutral-400 hover:text-white  backdrop-blur-sm transition-colors"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                </button>

                <div className="flex-1 p-2 bg-neutral-900 flex items-center justify-center min-h-[40vh] md:min-h-0">
                    <img
                        src={url ? url : imageUrl}
                        alt={photo.name || "Podgląd zdjęcia"}
                        className="w-full h-full max-h-[70vh] md:max-h-[85vh] object-contain select-none"
                    />
                </div>

                <div className="w-full md:w-80 p-6 flex flex-col justify-between border-t md:border-t-0 md:border-l border-neutral-800 bg-neutral-950/50 backdrop-blur-sm">
                    <div className="space-y-4">
                        <div>
                            <h2 className="text-xl font-bold text-neutral-50 tracking-tight">
                                {photo.name || "Bez tytułu"}
                            </h2>
                            {formattedDate && (
                                <div className="flex items-center gap-1.5 mt-1.5 text-xs text-neutral-500">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                                        <line x1="16" y1="2" x2="16" y2="6"></line>
                                        <line x1="8" y1="2" x2="8" y2="6"></line>
                                        <line x1="3" y1="10" x2="21" y2="10"></line>
                                    </svg>
                                    <time dateTime={photo.date}>{formattedDate}</time>
                                </div>
                            )}
                        </div>

                        {photo.description && (
                            <p className="text-sm text-neutral-400 leading-relaxed break-words">
                                {photo.description}
                            </p>
                        )}
                    </div>

                    {photo.location && (
                        <div className="mt-6 pt-4 border-t border-neutral-900 flex items-center gap-2 text-sm text-neutral-400">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-neutral-500 shrink-0">
                                <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
                                <circle cx="12" cy="10" r="3"></circle>
                            </svg>
                            <span className="truncate">{photo.location}</span>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}