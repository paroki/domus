import type { Route } from "./+types/home";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Domus" },
    { name: "description", content: "Parish Administration Digitalized" },
  ];
}

export default function Home() {
  return <h1>Domus: Parish Administration Digitalized</h1>;
}
