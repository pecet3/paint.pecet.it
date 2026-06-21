import PocketBase from 'pocketbase'
import type { RecordModel } from 'pocketbase';

export const pb = new PocketBase('https://my.pecet.it')

export interface PhotoRecord extends RecordModel {
    name: string;
    description: string;
    image: string;          // Nazwa pliku w PocketBase
    image_raw: string;      // Nazwa pliku raw
    location: string;
    date: string;           // ISO date string
    photo_collection: string;     // Relacja do innej kolekcji (RECORD_ID)
}

export interface PhotoCollectionRecord extends RecordModel {
    name: string;
    description: string;
    location: string;
    images: string[];       // Tablica nazw plików (w PocketBase dla pola "file" z obsługą wielu plików)
    date: string;           // ISO date string
}