import { getAuthToken } from "./auth"

export type TrackingPayload = {
    identifier: string,
    provider: string,
    trigger: string
}

export type Tracking = {
    image: string
    trigger: string
    pollSchedule: string
    provider: string
    namespace: string
    policy: string
    registry: string
}

const getTracking = async () => {
    const token = getAuthToken();
    const response = await fetch('/v1/tracked', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    return response.json();
}

const setTracking = async (payload: TrackingPayload) => {
    const token = getAuthToken();
    const response = await fetch('/v1/tracked', {
        method: 'PUT',
        headers: {
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(payload)
    });
    return response.json();
}

export {
    setTracking,
    getTracking
}