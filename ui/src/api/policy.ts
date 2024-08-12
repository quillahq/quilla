import { getAuthToken } from "./auth";

export type PolicyPayload = {
    identifier: string
    policy: string
    provider: string
}

const setPolicy = async (payload: PolicyPayload) => {
    const token = getAuthToken();
    const response = await fetch('/v1/policies', {
        method: 'PUT',
        headers: {
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(payload)
    });
    return response.json();
}

export {
    setPolicy
}