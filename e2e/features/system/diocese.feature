Feature: Diocese CRUD
  As an admin
  I want to manage dioceses in the system
  So that church territorial structures are properly maintained

  Scenario: Create a new diocese
    Given I am logged in as an admin
    And I open "/sys/diocese"
    When I click the add diocese button
    Then I should be on "/sys/diocese/create"
    When I fill the diocese name with "E2E Keuskupan Surabaya"
    And I submit the diocese form
    Then I should be on "/sys/diocese"
    And I should see a success notification "Keuskupan ditambahkan."
    And I should see diocese "E2E Keuskupan Surabaya" with status "Aktif" in the list

  Scenario: Create an inactive diocese and toggle active on edit
    Given I am logged in as an admin
    And I open "/sys/diocese"
    When I click the add diocese button
    Then I should be on "/sys/diocese/create"
    When I fill the diocese name with "E2E Keuskupan Pangkalpinang"
    And I set the diocese status to "Nonaktif"
    And I submit the diocese form
    Then I should be on "/sys/diocese"
    And I should see a success notification "Keuskupan ditambahkan."
    And I should see diocese "E2E Keuskupan Pangkalpinang" with status "Nonaktif" in the list
    When I click edit for diocese "E2E Keuskupan Pangkalpinang"
    And I set the diocese status to "Aktif"
    And I submit the diocese form
    Then I should be on "/sys/diocese"
    And I should see a success notification "Keuskupan diperbarui."
    And I should see diocese "E2E Keuskupan Pangkalpinang" with status "Aktif" in the list

  Scenario: Diocese form validation
    Given I am logged in as an admin
    And I open "/sys/diocese"
    When I click the add diocese button
    Then I should be on "/sys/diocese/create"
    When I submit the diocese form
    Then I should see a diocese validation error "Nama keuskupan wajib diisi."
    When I cancel the diocese form
    Then I should be on "/sys/diocese"

  Scenario: Edit an existing diocese
    Given I am logged in as an admin
    And I open "/sys/diocese"
    And diocese "E2E Keuskupan Bandung" exists
    And I refresh the page
    When I click edit for diocese "E2E Keuskupan Bandung"
    And I fill the diocese name with "E2E Keuskupan Bandung Updated"
    And I submit the diocese form
    Then I should be on "/sys/diocese"
    And I should see a success notification "Keuskupan diperbarui."
    And I should see diocese "E2E Keuskupan Bandung Updated" in the list
    And I should not see diocese "E2E Keuskupan Bandung" in the list

  Scenario: Cancel deletion keeps diocese
    Given I am logged in as an admin
    And I open "/sys/diocese"
    And diocese "E2E Keuskupan Bogor" exists
    And I refresh the page
    When I click delete for diocese "E2E Keuskupan Bogor"
    And I cancel the deletion
    Then I should see diocese "E2E Keuskupan Bogor" in the list

  Scenario: Delete a diocese
    Given I am logged in as an admin
    And I open "/sys/diocese"
    And diocese "E2E Keuskupan Denpasar" exists
    And I refresh the page
    When I click delete for diocese "E2E Keuskupan Denpasar"
    And I confirm the deletion
    Then I should see a success notification "Keuskupan dihapus."
    And I should not see diocese "E2E Keuskupan Denpasar" in the list

  Scenario: Search dioceses
    Given I am logged in as an admin
    And I open "/sys/diocese"
    And diocese "E2E Keuskupan Semarang" exists
    And diocese "E2E Keuskupan Malang" exists
    And I refresh the page
    When I search diocese for "Semarang"
    Then I should see diocese "E2E Keuskupan Semarang" in the list
    And I should not see diocese "E2E Keuskupan Malang" in the list

  Scenario: Non-admin cannot access dioceses
    Given I am logged in as a regular user
    When I open "/sys/diocese"
    Then I should see an error or forbidden message

