import { getAuthToken } from "./auth"

export type Audit = {
    id: string
    createdAt: string
    updatedAt: string
    accountId: string
    username: string
    email: string
    action: string
    resourceKind: string
    identifier: string
    message: string
    payload: string
    payloadType: string
    metadata: {
        [key: string]: string
    }
}

const getAudit = async () => {
    const token = getAuthToken();
    const response = await fetch('/v1/audit?filter=*&limit=0&offset=0', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    const json = await response.json()
    // @ts-ignore
    return json.data as Audit[];
}

export {
    getAudit
}