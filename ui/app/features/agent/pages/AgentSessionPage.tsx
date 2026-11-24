import { useParams } from "react-router";
import { AgentWorkspace } from "../components/AgentWorkspace";

export function meta() {
  return [{ title: "Asisten AI — Domus" }];
}

export default function AgentSessionPage() {
  const { sessionId } = useParams();
  return <AgentWorkspace initialSessionId={sessionId} />;
}
