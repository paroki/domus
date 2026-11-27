import type { components } from "~/shared/api/schema";

export type Cantus = components["schemas"]["domain.CantusResponse"];
export type CantusMedia = components["schemas"]["domain.CantusMedia"];
export type CantusSearchFilter =
  components["schemas"]["domain.CantusSearchFilter"];
export type SearchCantusRequest =
  components["schemas"]["domain.SearchCantusRequest"];
export type CreateCantusRequest =
  components["schemas"]["domain.CreateCantusRequest"];
export type UpdateCantusRequest =
  components["schemas"]["domain.UpdateCantusRequest"];
