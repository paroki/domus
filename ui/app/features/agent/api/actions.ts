import { useMutation, useQueryClient } from "@tanstack/react-query";
import { agentRequest } from "./http";
import { agentKeys } from "./keys";
import type {
  CreateMemoryRequest,
  MemoryResponse,
  SessionResponse,
  UpdateMemoryRequest,
  UpdateSessionRequest,
} from "./types";

export function useUpdateAgentSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      sessionId,
      body,
    }: {
      sessionId: string;
      body: UpdateSessionRequest;
    }) => {
      const res = await agentRequest<SessionResponse>(
        `/agent/sessions/${sessionId}`,
        {
          method: "PATCH",
          body: JSON.stringify(body),
        },
      );
      return res.data;
    },
    onSuccess: (data) => {
      void qc.invalidateQueries({ queryKey: agentKeys.session(data.id) });
      void qc.invalidateQueries({ queryKey: ["agent", "sessions"] });
    },
  });
}

export function useDeleteAgentSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (sessionId: string) => {
      await agentRequest<void>(`/agent/sessions/${sessionId}`, {
        method: "DELETE",
      });
      return sessionId;
    },
    onSuccess: (sessionId) => {
      void qc.removeQueries({ queryKey: agentKeys.session(sessionId) });
      void qc.invalidateQueries({ queryKey: ["agent", "sessions"] });
    },
  });
}

export function useCreateAgentMemory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: CreateMemoryRequest) => {
      const res = await agentRequest<MemoryResponse>("/agent/memories", {
        method: "POST",
        body: JSON.stringify(body),
      });
      return res.data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["agent", "memories"] });
    },
  });
}

export function useUpdateAgentMemory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      memoryId,
      body,
    }: {
      memoryId: string;
      body: UpdateMemoryRequest;
    }) => {
      const res = await agentRequest<MemoryResponse>(
        `/agent/memories/${memoryId}`,
        {
          method: "PATCH",
          body: JSON.stringify(body),
        },
      );
      return res.data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["agent", "memories"] });
    },
  });
}

export function useDeleteAgentMemory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (memoryId: string) => {
      await agentRequest<void>(`/agent/memories/${memoryId}`, {
        method: "DELETE",
      });
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["agent", "memories"] });
    },
  });
}
