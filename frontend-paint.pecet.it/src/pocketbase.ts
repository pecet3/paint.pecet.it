import PocketBase from 'pocketbase'
import type { RecordModel } from 'pocketbase';

export const pb = new PocketBase('https://my.pecet.it')

export interface Painting extends RecordModel {

}
