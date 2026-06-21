import { pb } from "../pocketbase";
import type { Photo } from "../Store";

export const PhotoView = ({ photo, onOpenModal }: { photo: Photo; onOpenModal?: (photo: Photo) => void }) => {
    const imageUrl = pb.files.getURL(photo, photo.image)

    return (
        <div
            onClick={() => onOpenModal && onOpenModal(photo)}
            className="bg-neutral-200 rounded-lg p-1.5 shadow-xl hover:ring-2 hover:shadow-lg ring-neutral-200 duration-500 flex flex-col gap-0.5 m-auto">
            <div

                className="group overflow-hidden rounded-sm  cursor-pointer
                 transition-all "
            >
                <img
                    src={imageUrl}
                    alt={photo.name}
                    className="w-full max-h-56  object-contain "
                    loading="lazy"
                />
            </div>
            <div className="">
                <h2 className="text-sm font-bold text-neutral-600 tracking-tight">{photo.name}</h2>
                {photo.description && (
                    <p className="text-neutral-400 text-sm leading-relaxed">{photo.description}</p>
                )}
                {photo.location && (
                    <span className="flex m-0  w-full justify-center  items-center  gap-1 text-[12px] text-neutral-500 ">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-neutral-500 shrink-0">
                            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
                            <circle cx="12" cy="10" r="3"></circle>
                        </svg>
                        {photo.location}
                    </span>
                )}
            </div>
        </div>
    )
}

export const PhotoViewSecondary = ({ photo, onOpenModal }: { photo: Photo; onOpenModal?: (photo: Photo) => void }) => {
    const imageUrl = pb.files.getURL(photo, photo.image)

    return (
        <div
            onClick={() => onOpenModal && onOpenModal(photo)}
            className="bg-neutral-200 rounded-lg p-1.5 shadow-xl hover:ring-2 hover:shadow-lg ring-neutral-200 duration-500 flex flex-col gap-0.5 m-auto">
            <div

                className="group overflow-hidden rounded-sm  cursor-pointer
                 transition-all "
            >
                <img
                    src={imageUrl}
                    alt={photo.name}
                    className="w-full h-56  object-fill "
                    loading="lazy"
                />
            </div>
            <div className="">
                <h2 className="text-sm font-bold text-neutral-600 tracking-tight">{photo.name}</h2>
                {photo.description && (
                    <p className="text-neutral-400 text-sm leading-relaxed">{photo.description}</p>
                )}
                {photo.location && (
                    <span className="flex m-0  w-full justify-center  items-center  gap-1 text-[12px] text-neutral-500 ">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-neutral-500 shrink-0">
                            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
                            <circle cx="12" cy="10" r="3"></circle>
                        </svg>
                        {photo.location}
                    </span>
                )}
            </div>
        </div>
    )
}