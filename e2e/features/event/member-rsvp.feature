@org-member
Feature: Organization Member Menu Visibility, Read-Only Directory, Event Visibility, and RSVP

  Scenario: Organization member sees scoped menu, read-only roster, restricted sibling roster, and can self-leave
    When I open "/"
    Then I see launcher tiles "Organisasi, Kegiatan"
    And I do not see launcher tiles "Pengaturan Paroki, Pengguna"
    When I open "/users"
    Then I see the users forbidden banner
    When I open my active organization detail
    Then I see organization detail without admin controls or invites
    And I see read-only org roster with member "Anggota Organisasi E2E"
    When I open the sibling organization detail
    Then I see the organization roster restricted banner
    When I open my active organization detail
    And I leave my active organization
    Then I am on "/org"

  Scenario: Organization member sees public events across parish and private events only for joined organization
    Given my organization has a "private" event "Latihan Internal Liturgi E2E" with RSVP enabled
    And the sibling organization has a "public" event "Seminar Publik Katekese E2E"
    And the sibling organization has a "private" event "Rapat Internal Katekese E2E"
    When I open "/event"
    Then I see the events page
    And I do not see the create event button
    And I see event card "Latihan Internal Liturgi E2E" with visibility "Internal Organisasi"
    And I see event card "Seminar Publik Katekese E2E" with visibility "Publik Paroki"
    And I do not see event card "Rapat Internal Katekese E2E"
    When I open event detail "Seminar Publik Katekese E2E"
    Then I see event detail "Seminar Publik Katekese E2E"
    And I see the RSVP restricted banner

  Scenario: Organization member submits, updates, removes RSVP, and sees closed state when event is cancelled
    Given my organization has a "private" event "Rapat Anggota Liturgi E2E" with RSVP enabled
    When I open "/event"
    And I open event detail "Rapat Anggota Liturgi E2E"
    Then I see event detail "Rapat Anggota Liturgi E2E"
    When I submit RSVP "Hadir" with note "Hadir tepat waktu bersama seksi koor"
    Then I see my RSVP status "Hadir" with note "Hadir tepat waktu bersama seksi koor"
    And I see RSVP summary counts going 1 maybe 0 notGoing 0
    And I see RSVP attendee "Anggota Organisasi E2E" with status "Hadir"
    When I submit RSVP "Mungkin" with note ""
    Then I see my RSVP status "Mungkin"
    And I see RSVP summary counts going 0 maybe 1 notGoing 0
    When I remove my RSVP
    Then I see my RSVP removed for "Anggota Organisasi E2E"
    When the organization admin cancels the seeded event via API
    Then I see event RSVP closed message
