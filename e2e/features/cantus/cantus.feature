Feature: Global Cantus catalog, cantus:admin delegation, 3-step wizard, and 5-section detail

  @super-admin
  Scenario: Super Admin without cantus:admin cannot mutate /cantus and delegates cantus:admin in /users
    Given a global user "cantus-admin-e2e@domus.test" named "Cantus Admin E2E" exists
    When I open "/cantus"
    Then I see the cantus directory without create CTA
    When I open "/cantus/new"
    Then I see the cantus forbidden card
    When I open "/users"
    And I open user detail for "cantus-admin-e2e@domus.test"
    And I grant cantus admin on user detail
    Then I see the cantus admin badge on user detail
    When I go back to the users directory
    Then I see user "cantus-admin-e2e@domus.test" with cantus admin badge in users directory

  @super-admin
  Scenario: cantus:admin creates and edits lagu and mazmur+BPI entries via 3-step wizard and 5-section detail
    Given I am signed in as cantus admin "cantus-admin-e2e@domus.test"
    When I open "/cantus"
    And I create a lagu entry via cantus wizard titled "Hai Makhluk Semua"
    Then I see the cantus detail page for "Hai Makhluk Semua" with inline media
    When I update cantus nada dasar to "E" and lirik to "Hai makhluk semua, nyanyikanlah pujian!"
    Then I see "E" and "Hai makhluk semua, nyanyikanlah pujian!" on cantus detail
    When I create a mazmur entry via cantus wizard titled "Mazmur Minggu Adven I"
    Then I see the mazmur refrain, bait, and BPI blocks on cantus detail

  @super-admin
  Scenario: Cross-parish read-only search, deactivation, soft-delete, and cantus:admin revocation
    Given I am signed in as cantus admin "cantus-admin-e2e@domus.test"
    When I sign in as regular global user "cantus-reader-e2e@domus.test"
    And I open "/cantus"
    And I search cantus directory for "Makhluk" with sumber "PS" and tag "Adven"
    Then I see cantus song "Hai Makhluk Semua" in the directory
    When I open cantus detail for "Hai Makhluk Semua"
    Then I see cantus detail in read-only mode
    When I sign in as cantus admin "cantus-admin-e2e@domus.test"
    And I open "/cantus"
    And I open cantus detail for "Hai Makhluk Semua"
    And I deactivate and soft-delete the cantus entry
    Then I do not see cantus song "Hai Makhluk Semua" in the directory
    When I sign in as super admin
    And I open "/users"
    And I open user detail for "cantus-admin-e2e@domus.test"
    And I revoke cantus admin on user detail
    Then I do not see the cantus admin badge on user detail
