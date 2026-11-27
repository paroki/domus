Feature: Paroki Admin parish-level events lifecycle

  @paroki-admin
  Scenario: Paroki Admin creates, filters, edits, and deletes a parish-level event with paroki timezone
    When I open "/event"
    Then I see the events page
    When I create a parish-level event with title "Misa Natal Paroki E2E" start "2099-12-25T08:00" end "2099-12-25T10:00" location "Gereja Katedral"
    Then I see event card "Misa Natal Paroki E2E" with visibility "Publik Paroki"
    And I see event card "Misa Natal Paroki E2E" with timezone "WITA"
    When I filter events by organizer "Paroki (Tingkat Paroki)"
    Then I see event card "Misa Natal Paroki E2E" with visibility "Publik Paroki"
    When I open event detail "Misa Natal Paroki E2E"
    Then I see event detail "Misa Natal Paroki E2E"
    And I see event detail timezone "WITA"
    When I edit the event jenis to "Misa / Ekaristi"
    Then I see event liturgi shortcut button
    When I click event liturgi shortcut button
    Then I see liturgi detail "Misa Natal Paroki E2E"
    When I open "/event"
    And I open event detail "Misa Natal Paroki E2E"
    When I edit the event to title "Misa Natal Agung Paroki E2E" location "Gereja Utama" visibility "Publik Paroki"
    Then I see event detail "Misa Natal Agung Paroki E2E"
    When I delete the event
    Then I am on "/event"
    And I do not see event card "Misa Natal Agung Paroki E2E"
