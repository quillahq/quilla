import { getAuthToken } from "./auth"

export type Stat = {
    date: string
    webhooks: number
    approved: number
    rejected: number
    updates: number
}

const getStats = async () => {
    const token = getAuthToken();
    const response = await fetch('/v1/stats', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    return response.json();
}

export {
    getStats
}