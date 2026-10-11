Feature: Launcher

  Background:
    Given I am logged in as an admin

  Scenario: Show all modules
    When I open "/"
    Then I should see the launcher
    And I should see modules "Sistem, Website, Sakramen, Keuangan, Umat, Kegiatan"

  Scenario: Open a module
    When I open "/"
    And I click module "Keuangan"
    Then I should be on "/fin"

