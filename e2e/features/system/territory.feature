Feature: Territory CRUD
  As an admin
  I want to manage territorial areas in the system
  So that church neighborhood and regional structures are properly maintained

  Scenario: Create a new root territory
    Given I am logged in as an admin
    And parish "E2E Paroki Wilayah Test" exists
    And I open "/sys/territorial"
    When I click the add territory button
    Then I should be on "/sys/territorial/create"
    When I fill the territory name with "E2E Wilayah Santo Petrus"
    And I fill the territory code with "WIL-01"
    And I select territory kind "Wilayah" in territory form
    And I select parish "E2E Paroki Wilayah Test" in territory form
    And I fill the territory description with "Wilayah satu paroki test"
    And I submit the territory form
    Then I should be on "/sys/territorial"
    And I should see a success notification "Wilayah teritorial ditambahkan."
    And I should see territory "E2E Wilayah Santo Petrus" with kind "Wilayah" in the list
    And I should see territory "E2E Wilayah Santo Petrus" with status "Aktif" in the list

  Scenario: Create child territory under parent
    Given I am logged in as an admin
    And parish "E2E Paroki Hierarki" exists
    And territory "E2E Wilayah Induk" exists in parish "E2E Paroki Hierarki" with kind "region"
    And I open "/sys/territorial"
    When I click the add territory button
    Then I should be on "/sys/territorial/create"
    When I fill the territory name with "E2E Lingkungan Anak"
    And I fill the territory code with "LING-01"
    And I select territory kind "Lingkungan" in territory form
    And I select parish "E2E Paroki Hierarki" in territory form
    And I select parent territory "E2E Wilayah Induk" in territory form
    And I submit the territory form
    Then I should be on "/sys/territorial"
    And I should see a success notification "Wilayah teritorial ditambahkan."
    And I should see territory "E2E Lingkungan Anak" with kind "Lingkungan" in the list

  Scenario: Territory form validation
    Given I am logged in as an admin
    And I open "/sys/territorial"
    When I click the add territory button
    Then I should be on "/sys/territorial/create"
    When I submit the territory form
    Then I should see a territory validation error "Nama wilayah wajib diisi."
    And I should see a territory validation error "Jenis wilayah wajib dipilih."
    And I should see a territory validation error "Paroki wajib dipilih."
    When I cancel the territory form
    Then I should be on "/sys/territorial"

  Scenario: Edit an existing territory
    Given I am logged in as an admin
    And parish "E2E Paroki Edit" exists
    And territory "E2E Wilayah To Edit" exists in parish "E2E Paroki Edit" with kind "region"
    And I open "/sys/territorial"
    When I click edit for territory "E2E Wilayah To Edit"
    And I fill the territory name with "E2E Wilayah Edited"
    And I fill the territory code with "WIL-EDITED"
    And I submit the territory form
    Then I should be on "/sys/territorial"
    And I should see a success notification "Wilayah teritorial diperbarui."
    And I should see territory "E2E Wilayah Edited" in the list
    And I should not see territory "E2E Wilayah To Edit" in the list

  Scenario: Delete a territory
    Given I am logged in as an admin
    And parish "E2E Paroki Delete" exists
    And territory "E2E Wilayah To Delete" exists in parish "E2E Paroki Delete" with kind "station"
    And I open "/sys/territorial"
    When I click delete for territory "E2E Wilayah To Delete"
    And I confirm the deletion
    Then I should see a success notification "Wilayah teritorial dihapus."
    And I should not see territory "E2E Wilayah To Delete" in the list

  Scenario: Search territory by name
    Given I am logged in as an admin
    And parish "E2E Paroki Search" exists
    And territory "E2E Wilayah Alfa" exists in parish "E2E Paroki Search" with kind "region"
    And territory "E2E Wilayah Beta" exists in parish "E2E Paroki Search" with kind "region"
    And I open "/sys/territorial"
    When I search territory for "Alfa"
    Then I should see territory "E2E Wilayah Alfa" in the list
    And I should not see territory "E2E Wilayah Beta" in the list

