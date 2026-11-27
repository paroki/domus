import type { components } from "~/shared/api/schema";

export type EventResponse = components["schemas"]["domain.EventResponse"];
export type SearchEventsRequest =
  components["schemas"]["domain.SearchEventsRequest"];
export type EventSearchFilter =
  components["schemas"]["domain.EventSearchFilter"];
export type CreateEventRequest =
  components["schemas"]["domain.CreateEventRequest"];
export type UpdateEventRequest =
  components["schemas"]["domain.UpdateEventRequest"];

export type EventRSVPResponse =
  components["schemas"]["domain.EventRSVPResponse"];
export type UpsertEventRSVPRequest =
  components["schemas"]["domain.UpsertEventRSVPRequest"];
export type EventRSVPSummaryResponse =
  components["schemas"]["domain.EventRSVPSummaryResponse"];
export type SearchEventRSVPsRequest =
  components["schemas"]["domain.SearchEventRSVPsRequest"];
export type EventRSVPSearchFilter =
  components["schemas"]["domain.EventRSVPSearchFilter"];

export type OrganizationResponse =
  components["schemas"]["domain.OrganizationResponse"];
