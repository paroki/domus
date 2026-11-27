Feature: Paroki Admin settings, profile, wilayah, and lookups

  @paroki-admin
  Scenario: Paroki Admin sees paroki settings tiles and is denied super-admin paroki route
    When I open "/settings"
    Then I see the settings page
    And I see settings tiles "Administrasi Paroki, Kamus Referensi"
    And I do not see settings tiles "Struktur Gerejawi, Wilayah Sipil, Sistem & Integrasi"
    When I open "/settings/gerejawi/paroki"
    Then I see the super admin restriction banner on paroki page

  @paroki-admin
  Scenario: Paroki Admin updates paroki profile and changes persist across reload
    When I open "/settings/paroki/profile"
    Then I see the paroki profile page
    When I update the paroki profile with:
      | name        | Paroki St. Yohanes E2E |
      | patronSaint | St. Yohanes Rasul      |
      | pastorName  | RD. Antonius E2E       |
      | phone       | +6281234567899         |
      | email       | sekretariat@e2e.org    |
      | website     | https://e2e.paroki.org |
      | address     | Jl. Gereja No. 1       |
      | timezone    | Asia/Jakarta           |
      | latitude    | -6.1695                |
      | longitude   | 106.8331               |
    Then I see the profile saved banner
    When I open "/settings/paroki/profile"
    Then I see the paroki profile fields match:
      | name        | Paroki St. Yohanes E2E |
      | patronSaint | St. Yohanes Rasul      |
      | pastorName  | RD. Antonius E2E       |
      | phone       | +6281234567899         |
      | email       | sekretariat@e2e.org    |
      | website     | https://e2e.paroki.org |
      | address     | Jl. Gereja No. 1       |
      | timezone    | Asia/Jakarta           |

  @paroki-admin
  Scenario: Paroki Admin manages wilayah and lingkungan with REF_IN_USE delete guard
    When I open "/settings/paroki/wilayah"
    Then I see the wilayah page
    When I create a wilayah named "Wilayah Auto E2E" expecting auto code "wilayah-auto-e2e"
    Then I see wilayah "Wilayah Auto E2E" with code "wilayah-auto-e2e"
    When I filter wilayah by status "false"
    Then I do not see wilayah "Wilayah Auto E2E"
    When I filter wilayah by status "true"
    And I create a lingkungan under wilayah "Wilayah Auto E2E" named "Lingkungan Auto E2E" expecting auto code "lingkungan-auto-e2e"
    Then I see lingkungan "Lingkungan Auto E2E" with code "lingkungan-auto-e2e"
    When I filter lingkungan by status "false"
    Then I do not see lingkungan "Lingkungan Auto E2E"
    When I filter lingkungan by status "true"
    And I delete wilayah "Wilayah Auto E2E"
    Then I see wilayah error "Data masih digunakan oleh data lain dan tidak dapat dihapus."
    When I delete lingkungan "Lingkungan Auto E2E"
    Then I do not see lingkungan "Lingkungan Auto E2E"
    When I delete wilayah "Wilayah Auto E2E"
    Then I do not see wilayah "Wilayah Auto E2E"

  @paroki-admin
  Scenario: Paroki Admin views read-only global lookups and manages paroki-scoped lookups
    When I open "/settings/lookups"
    Then I do not see lookup category "Agama" on settings lookups
    And I do not see lookup category "Golongan Darah" on settings lookups
    Given a global lookup exists in category "status_paroki" with code "global-ro-e2e" and label "Status Global E2E"
    When I open "/settings/lookups?category=status_paroki"
    Then I see the lookups page
    And I see lookup "Status Global E2E" with scope "Global"
    And lookup "Status Global E2E" is read-only
    When I create a paroki lookup in category "org_position" with code "sekretaris-e2e" and label "Sekretaris E2E"
    Then I see lookup "Sekretaris E2E" with scope "Paroki"
    When I edit lookup "Sekretaris E2E" to label "Sekretaris E2E Updated"
    Then I see lookup "Sekretaris E2E Updated" with scope "Paroki"
    When I toggle active state for lookup "Sekretaris E2E Updated"
    And I filter lookups by status "true"
    Then I do not see lookup "Sekretaris E2E Updated"
    When I filter lookups by status "false"
    Then I see lookup "Sekretaris E2E Updated" with scope "Paroki"
    When I delete lookup "Sekretaris E2E Updated"
    Then I do not see lookup "Sekretaris E2E Updated"
