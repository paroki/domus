Feature: Invite-only login guard, guest invite sign-up, onboarding, and profile security

  Scenario: Uninvited email is rejected on /login
    When I open "/login"
    Then I see the login page
    When I request email OTP on login for "stranger.uninvited@domus.test"
    Then I see login error containing "Akun tidak ditemukan"

  Scenario: New guest signs up via /org/join/:code, completes onboarding, and lands on /org/:orgId
    Given an organization "Paduan Suara E2E" has an active member invite
    When I open the organization invite link
    Then I see invite preview for "Paduan Suara E2E" with role "Anggota"
    When I request guest OTP with name "Tamu Undangan E2E" and email "tamu.undangan.e2e@domus.test"
    And a deterministic OTP "123456" is seeded for "tamu.undangan.e2e@domus.test"
    And I verify guest OTP "123456" for "Paduan Suara E2E"
    Then I am redirected to onboarding
    And I see onboarding birth date bounds
    When I complete onboarding with name "Tamu Undangan Terverifikasi", phone "+6281298765432", gender "wanita", birth date "1996-08-17", and timezone "Asia/Jakarta"
    Then I am on the invited organization detail page "Paduan Suara E2E"
    And I see read-only org roster with member "Tamu Undangan Terverifikasi"
    When I open "/me"
    Then I see the profile and security page
    And I see the device and notification card on profile
    When I toggle notification channel "Notifikasi Push" off
    Then notification channel "Notifikasi Push" is unchecked
    When I toggle notification channel "Notifikasi Push" on
    Then notification channel "Notifikasi Push" is checked
    When I click test notification button
    Then I see test notification sent message
    And I see my profile shows gender "wanita", birth date "1996-08-17", and timezone "Asia/Jakarta"
    When I update my profile to name "Tamu Undangan Terverifikasi", phone "+6281298765432", gender "pria", birth date "1988-04-12", and timezone "Asia/Jayapura"
    Then I see profile saved message
    And I see my profile shows gender "pria", birth date "1988-04-12", and timezone "Asia/Jayapura"

  Scenario: Revoked invite link is rejected on /org/join/:code
    Given an organization "Paduan Suara Revoked E2E" has a revoked invite
    When I open the organization invite link
    Then I see revoked invite message "Tautan undangan ini berstatus Dicabut dan tidak dapat digunakan lagi."

  @authenticated
  Scenario: Authenticated user updates profile, revokes secondary session, verifies last account guard, and logs out
    Given a secondary session exists for "user@domus.test"
    When I open "/me"
    Then I see the profile and security page
    And I see the device and notification card on profile
    When I toggle notification channel "Notifikasi Push" off
    Then notification channel "Notifikasi Push" is unchecked
    When I toggle notification channel "Notifikasi Push" on
    Then notification channel "Notifikasi Push" is checked
    When I click test notification button
    Then I see test notification sent message
    When I update my profile to name "Umat Mandiri E2E", phone "+6281399998888", gender "pria", and birth date "1992-03-14"
    Then I see profile saved message
    And I see my profile shows gender "pria" and birth date "1992-03-14"
    And I see current device session in active sessions
    When I revoke the secondary session
    Then I see last linked account protection
    When I create a personal access token named "CLI Exporter E2E" with expiry "30"
    Then I see the one-time personal access token banner and token "CLI Exporter E2E" in the list
    And I can authenticate GET "/api/me" with the personal access token
    When I revoke the personal access token "CLI Exporter E2E"
    Then the personal access token "CLI Exporter E2E" is rejected by "/api/me"
    When I log out from header menu
    Then I am on "/login"

  @paroki-admin
  Scenario: Unverified user completes KTP-only self-verification and sees verified Data Umat card on /me
    Given a paroki member "verify-ktp-e2e@domus.test" named "Verify KTP E2E" exists in the active paroki
    And I sign in as "verify-ktp-e2e@domus.test" on the active paroki
    When I open "/me"
    Then I see the verification CTA card on profile
    When I complete KTP-only self-verification for "Verify KTP E2E" with nik "3174010101901111"
    Then I am on "/me"
    And I see the verified Data Umat card on profile

  @paroki-admin
  Scenario: Unverified user completes Kartu Pelajar and KK self-verification with member row selection
    Given a paroki member "verify-kk-e2e@domus.test" named "Pelajar E2E" exists in the active paroki
    And I sign in as "verify-kk-e2e@domus.test" on the active paroki
    When I complete Kartu Pelajar and KK self-verification selecting "Pelajar E2E" with nomorKk "3174010101902222"
    Then I am on "/me"
    And I see the verified Data Umat card on profile

  @paroki-admin
  Scenario: Self-verification shows conflict error when umat is already linked
    Given a paroki member "verify-conflict-e2e@domus.test" named "Verify Conflict E2E" exists in the active paroki
    And I sign in as "verify-conflict-e2e@domus.test" on the active paroki
    When I submit KTP self-verification that hits conflict "UMAT_USER_ALREADY_LINKED"
    Then I see verification error containing "sudah terhubung dengan akun Domus lain"
