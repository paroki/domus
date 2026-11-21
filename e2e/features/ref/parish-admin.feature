Feature: Parish Admin parish settings, profile, territory, lookups, and periods

  @parish-admin
  Scenario: Parish Admin sees parish settings tiles and is denied super-admin parishes route
    When I open "/settings"
    Then I see the settings page
    And I see settings tiles "Profil Paroki, Wilayah & Lingkungan, Kamus Referensi, Periode"
    And I do not see settings tiles "Keuskupan & Paroki"
    When I open "/settings/parishes"
    Then I see the super admin restriction banner on parishes page

  @parish-admin
  Scenario: Parish Admin updates parish profile and changes persist across reload
    When I open "/settings/profile"
    Then I see the parish profile page
    When I update the parish profile with:
      | name        | Paroki St. Yohanes E2E |
      | patronSaint | St. Yohanes Rasul      |
      | pastorName  | RD. Antonius E2E       |
      | phone       | +6281234567899         |
      | email       | sekretariat@e2e.org    |
      | website     | https://e2e.paroki.org |
      | address     | Jl. Gereja No. 1       |
    Then I see the profile saved banner
    When I open "/settings/profile"
    Then I see the parish profile fields match:
      | name        | Paroki St. Yohanes E2E |
      | patronSaint | St. Yohanes Rasul      |
      | pastorName  | RD. Antonius E2E       |
      | phone       | +6281234567899         |
      | email       | sekretariat@e2e.org    |
      | website     | https://e2e.paroki.org |
      | address     | Jl. Gereja No. 1       |

  @parish-admin
  Scenario: Parish Admin manages zones and wards with REF_IN_USE delete guard
    When I open "/settings/territory"
    Then I see the territory page
    When I create a zone with code "wil-e2e-1" and name "Wilayah E2E 1"
    Then I see zone "Wilayah E2E 1" with code "wil-e2e-1"
    When I create a ward under zone "Wilayah E2E 1" with code "ling-e2e-1" and name "Lingkungan E2E 1"
    Then I see ward "Lingkungan E2E 1" with code "ling-e2e-1"
    When I delete zone "Wilayah E2E 1"
    Then I see territory error "Data masih digunakan oleh data lain dan tidak dapat dihapus."
    When I delete ward "Lingkungan E2E 1"
    Then I do not see ward "Lingkungan E2E 1"
    When I delete zone "Wilayah E2E 1"
    Then I do not see zone "Wilayah E2E 1"

  @parish-admin
  Scenario: Parish Admin views read-only global lookups and manages parish-scoped lookups
    Given a global lookup exists in category "sacrament" with code "global-ro-e2e" and label "Sakramen Global E2E"
    When I open "/settings/lookups"
    Then I see the lookups page
    And I see lookup "Sakramen Global E2E" with scope "Global"
    And lookup "Sakramen Global E2E" is read-only
    When I create a parish lookup in category "org_position" with code "sekretaris-e2e" and label "Sekretaris E2E"
    Then I see lookup "Sekretaris E2E" with scope "Paroki"
    When I edit lookup "Sekretaris E2E" to label "Sekretaris E2E Updated"
    Then I see lookup "Sekretaris E2E Updated" with scope "Paroki"
    When I toggle active state for lookup "Sekretaris E2E Updated"
    And I delete lookup "Sekretaris E2E Updated"
    Then I do not see lookup "Sekretaris E2E Updated"

  @parish-admin
  Scenario: Parish Admin manages fiscal periods with overlap guard and active rotation
    When I open "/settings/periods"
    Then I see the periods page
    When I create a fiscal period with code "fy-2025-e2e" name "Tahun Buku 2025 E2E" from "2025-01-01" to "2025-12-31" as current
    Then I see period "Tahun Buku 2025 E2E" with status "Terbuka" and current badge
    When I create a fiscal period with code "fy-overlap-e2e" name "Tahun Buku Overlap" from "2025-06-01" to "2026-06-01"
    Then I see period error "Rentang tanggal periode bertumpuk dengan periode lain yang sudah ada."
    When I create a fiscal period with code "fy-2026-e2e" name "Tahun Buku 2026 E2E" from "2026-01-01" to "2026-12-31"
    Then I see period "Tahun Buku 2026 E2E" with status "Terbuka"
    When I make period "Tahun Buku 2026 E2E" current
    Then I see period "Tahun Buku 2026 E2E" with status "Terbuka" and current badge
    And I see period "Tahun Buku 2025 E2E" without current badge
    When I close period "Tahun Buku 2025 E2E"
    Then I see period "Tahun Buku 2025 E2E" with status "Ditutup"
    When I delete period "Tahun Buku 2026 E2E"
    Then I do not see period "Tahun Buku 2026 E2E"
