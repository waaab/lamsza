import { hoursConfigured } from "./entryHours.js";

export function showListingPhotos(entry) {
    return Boolean(entry?.claimed) && Array.isArray(entry?.photos) && entry.photos.length > 0;
}
export function showListingHours(entry) {
    return Boolean(entry?.claimed) && hoursConfigured(entry?.hours);
}
export function showListingDeliveryHours(entry) {
    return Boolean(entry?.claimed) && hoursConfigured(entry?.delivery_hours);
}
export function showListingRatings(entry) {
    return Boolean(entry?.claimed) && Boolean(entry?.ratings_enabled);
}
export function showListingTodayHours(entry) {
    return showListingHours(entry);
}
