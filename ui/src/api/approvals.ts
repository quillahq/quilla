import { getAuthToken } from "./auth";

export type Approval = {
  createdAt: string;
  currentVersion: string;
  deadline: string;
  digest: string;
  event: {
    createdAt: string;
    repository: {
      digest: string;
      host: string;
      name: string;
      tag: string;
    };
  };
  id: string;
  identifier: string;
  message: string;
  newVersion: string;
  provider: string;
  updatedAt: string;
  voters: any;
  rejected: boolean;
  archived: boolean;
  votesReceived: number;
  votesRequired: number;
};

export type ApprovalPayload = {
  identifier: string;
  provider: string;
  votesRequired: number;
};

export type ApprovalVotePayload = {
  id: string;
  identifier: string;
  action: string;
};

const voteApproval = async (payload: ApprovalVotePayload) => {
  const token = getAuthToken();
  const response = await fetch("/v1/approvals", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  });
  return response.json();
};

const setApproval = async (payload: ApprovalPayload) => {
  const token = getAuthToken();
  const response = await fetch("/v1/approvals", {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  });
  return response.json();
};

const getApprovals = async () => {
  const token = getAuthToken();
  const response = await fetch("/v1/approvals", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  return response.json();
};

const approvalIsComplete = (approval: Approval) =>
  approval.archived ||
  approval.rejected ||
  approval.votesReceived >= approval.votesRequired;

export { getApprovals, setApproval, voteApproval, approvalIsComplete };
