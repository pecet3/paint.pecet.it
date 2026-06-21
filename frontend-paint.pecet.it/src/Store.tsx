import React, { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import { pb, type PhotoRecord, type PhotoCollectionRecord } from './pocketbase';


export type Photo = PhotoRecord
export type PhotoCollection = PhotoCollectionRecord & {
    photos: Photo[] | PhotoRecord[] | null;
}
export type PhotosFeedItem = Photo | PhotoCollection;

interface StoreContextType {
    photosFeed: PhotosFeedItem[];
    loading: boolean;
    error: string | null;
    photoModal: Photo | null;
    setPhotoModal: (photo: Photo | null) => void;
    fetchPhotosPageFeed: () => Promise<void>;
    fetchSinglePhoto: (id: string) => Promise<PhotoRecord | null>;
    fetchSingleCollection: (id: string) => Promise<PhotoCollectionRecord | null>;
}

const StoreContext = createContext<StoreContextType | undefined>(undefined);

interface StoreProviderProps {
    children: ReactNode;
}

export const isPhotoType = (item: PhotosFeedItem): boolean => {
    if ('image' in item) {
        return true
    }
    return false
}
export const StoreProvider: React.FC<StoreProviderProps> = ({ children }) => {
    const [photosFeed, setPhotosFeed] = useState<PhotosFeedItem[]>([]);
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);
    const [photoModal, setPhotoModal] = useState<Photo | null>(null)
    const fetchPhotosPageFeed = async () => {
        setLoading(true);
        setError(null);
        try {
            const [rPhotos, rCollections] = await Promise.all([
                pb.collection('photos').getFullList<PhotoRecord>({
                    sort: '-date',
                }),
                pb.collection('photo_collections').getFullList<PhotoCollectionRecord>({
                    sort: '-date',
                })
            ]);
            console.log(rPhotos, rCollections)
            const collectionsMap = new Map<string, PhotoCollection>();
            rCollections.forEach(coll => {
                collectionsMap.set(coll.id, {
                    ...coll,
                    photos: []
                });
            });
            const standalonePhotos: Photo[] = [];
            rPhotos.forEach(photo => {
                const collectionId = photo.photo_collection;

                if (collectionId && collectionId.trim() !== "") {
                    const targetCollection = collectionsMap.get(collectionId);
                    if (targetCollection) {
                        targetCollection.photos?.push(photo);
                    } else {
                        standalonePhotos.push(photo);
                    }
                } else {
                    standalonePhotos.push(photo);
                }
            });

            const mappedCollections = Array.from(collectionsMap.values());

            mappedCollections.forEach(coll => {
                if (coll.photos && coll.photos.length === 0) {
                    coll.photos = null;
                }
            });

            const combinedFeed: PhotosFeedItem[] = [...standalonePhotos, ...mappedCollections].sort(
                (a, b) => new Date(b.created).getTime() - new Date(a.created).getTime()
            );
            console.log(combinedFeed)
            setPhotosFeed(combinedFeed);


        } catch (err: any) {
            setError(err.message || 'Error fetching feed data.');
        } finally {
            setLoading(false);
        }
    };

    const fetchSinglePhoto = async (id: string): Promise<PhotoRecord | null> => {
        try {
            return await pb.collection('photos').getOne<PhotoRecord>(id);
        } catch {
            return null;
        }
    };

    const fetchSingleCollection = async (id: string): Promise<PhotoCollectionRecord | null> => {
        try {
            return await pb.collection('photo_collections').getOne<PhotoCollectionRecord>(id);
        } catch {
            return null;
        }
    };

    useEffect(() => {
        fetchPhotosPageFeed();
    }, []);

    return (
        <StoreContext.Provider value={{
            photosFeed,
            loading,
            error,
            photoModal,
            setPhotoModal,
            fetchPhotosPageFeed,
            fetchSinglePhoto,
            fetchSingleCollection
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