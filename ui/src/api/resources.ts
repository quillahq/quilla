import { getAuthToken } from "./auth";

export type Resource = {
    provider: string;
    identifier: string;
    name: string;
    namespace: string;
    kind: string;
    policy: string;
    images: string[];
    labels: {
      [key: string]: string;
    };
    // annotations: {
    //     [key: string]: string
    // }
    status: {
      replicas: number;
      updatedReplicas: number;
      readyReplicas: number;
      availableReplicas: number;
      unavailableReplica: number;
    };
  }

const getResources = async () => {
    const token = getAuthToken();
    const response = await fetch('/v1/resources', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    return response.json();
}

export {
    getResources,
}