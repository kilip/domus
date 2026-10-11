Feature: Parish CRUD
  As an admin
  I want to manage parishes in the system
  So that church territorial structures are properly maintained

  Scenario: Create a new parish
    Given I am logged in as an admin
    And I open "/sys/parish"
    When I click the add parish button
    Then I should be on "/sys/parish/create"
    When I fill the parish name with "E2E Paroki Katedral Surabaya"
    And I fill the parish church with "Hati Kudus Yesus"
    And I fill the parish address with "Jl. Polisi Istimewa No. 15, Surabaya"
    And I fill the parish phone with "031-5678901"
    And I fill the parish email with "katedral.sby@example.com"
    And I submit the parish form
    Then I should be on "/sys/parish"
    And I should see a success notification "Paroki ditambahkan."
    And I should see parish "E2E Paroki Katedral Surabaya" with church "Hati Kudus Yesus" in the list
    And I should see parish "E2E Paroki Katedral Surabaya" with status "Aktif" in the list

  Scenario: Create an inactive parish and toggle active on edit
    Given I am logged in as an admin
    And I open "/sys/parish"
    When I click the add parish button
    Then I should be on "/sys/parish/create"
    When I fill the parish name with "E2E Paroki Nonaktif"
    And I set the parish status to "Nonaktif"
    And I submit the parish form
    Then I should be on "/sys/parish"
    And I should see a success notification "Paroki ditambahkan."
    And I should see parish "E2E Paroki Nonaktif" with status "Nonaktif" in the list
    When I click edit for parish "E2E Paroki Nonaktif"
    And I set the parish status to "Aktif"
    And I submit the parish form
    Then I should be on "/sys/parish"
    And I should see a success notification "Paroki diperbarui."
    And I should see parish "E2E Paroki Nonaktif" with status "Aktif" in the list

  Scenario: Parish form validation
    Given I am logged in as an admin
    And I open "/sys/parish"
    When I click the add parish button
    Then I should be on "/sys/parish/create"
    When I submit the parish form
    Then I should see a parish validation error "Nama paroki wajib diisi."
    When I cancel the parish form
    Then I should be on "/sys/parish"

  Scenario: Create a parish with diocese
    Given I am logged in as an admin
    And diocese "E2E Keuskupan Malang" exists
    And I open "/sys/parish"
    When I click the add parish button
    Then I should be on "/sys/parish/create"
    When I fill the parish name with "E2E Paroki Ijen"
    And I select diocese "E2E Keuskupan Malang" in parish form
    And I submit the parish form
    Then I should be on "/sys/parish"
    And I should see a success notification "Paroki ditambahkan."
    And I should see parish "E2E Paroki Ijen" in the list

  Scenario: Edit an existing parish
    Given I am logged in as an admin
    And I open "/sys/parish"
    And parish "E2E Paroki Santo Yusup" exists
    And I refresh the page
    When I click edit for parish "E2E Paroki Santo Yusup"
    And I fill the parish name with "E2E Paroki Santo Yusup Updated"
    And I fill the parish church with "Santo Yusup Pekerja"
    And I fill the parish address with "Jl. Gedangan No. 10"
    And I fill the parish phone with "024-3541624"
    And I fill the parish email with "yusup.updated@example.com"
    And I submit the parish form
    Then I should be on "/sys/parish"
    And I should see a success notification "Paroki diperbarui."
    And I should see parish "E2E Paroki Santo Yusup Updated" with church "Santo Yusup Pekerja" in the list
    And I should not see parish "E2E Paroki Santo Yusup" in the list

  Scenario: Cancel deletion keeps parish
    Given I am logged in as an admin
    And I open "/sys/parish"
    And parish "E2E Paroki Theresia" exists
    And I refresh the page
    When I click delete for parish "E2E Paroki Theresia"
    And I cancel the deletion
    Then I should see parish "E2E Paroki Theresia" in the list

  Scenario: Delete a parish
    Given I am logged in as an admin
    And I open "/sys/parish"
    And parish "E2E Paroki Mikael" exists
    And I refresh the page
    When I click delete for parish "E2E Paroki Mikael"
    And I confirm the deletion
    Then I should see a success notification "Paroki dihapus."
    And I should not see parish "E2E Paroki Mikael" in the list

  Scenario: Search parishes by name and church
    Given I am logged in as an admin
    And I open "/sys/parish"
    And parish "E2E Paroki Arnoldus" with church "Santo Arnoldus Janssen" exists
    And parish "E2E Paroki Fransiskus" with church "Santo Fransiskus Asisi" exists
    And I refresh the page
    When I search parish for "Arnoldus"
    Then I should see parish "E2E Paroki Arnoldus" in the list
    And I should not see parish "E2E Paroki Fransiskus" in the list
    When I search parish for "Asisi"
    Then I should see parish "E2E Paroki Fransiskus" in the list
    And I should not see parish "E2E Paroki Arnoldus" in the list

  Scenario: Non-admin cannot access parishes
    Given I am logged in as a regular user
    When I open "/sys/parish"
    Then I should see an error or forbidden message

