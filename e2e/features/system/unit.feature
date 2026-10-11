Feature: Unit CRUD
  As an admin
  I want to manage organizational units in the system
  So that parish councils, sections, committees, and groups are properly maintained

  Scenario: Create a new standard unit
    Given I am logged in as an admin
    And parish "E2E Paroki Unit Test" exists
    And I open "/sys/unit"
    When I click the add unit button
    Then I should be on "/sys/unit/create"
    When I fill the unit name with "E2E Dewan Paroki Harian"
    And I fill the unit code with "DPH-01"
    And I select unit kind "Dewan" in unit form
    And I select parish "E2E Paroki Unit Test" in unit form
    And I fill the unit description with "Dewan paroki harian periode aktif"
    And I submit the unit form
    Then I should be on "/sys/unit"
    And I should see a success notification "Unit organisasi ditambahkan."
    And I should see unit "E2E Dewan Paroki Harian" with kind "Dewan" in the list
    And I should see unit "E2E Dewan Paroki Harian" with status "Aktif" in the list

  Scenario: Create territorial unit linked to territory
    Given I am logged in as an admin
    And parish "E2E Paroki Unit Teritorial" exists
    And territory "E2E Wilayah Unit Test" exists in parish "E2E Paroki Unit Teritorial" with kind "region"
    And I open "/sys/unit"
    When I click the add unit button
    Then I should be on "/sys/unit/create"
    When I fill the unit name with "E2E Pengurus Wilayah Satu"
    And I fill the unit code with "PW-01"
    And I select unit kind "Teritorial" in unit form
    And I select parish "E2E Paroki Unit Teritorial" in unit form
    And I select territory "E2E Wilayah Unit Test" in unit form
    And I submit the unit form
    Then I should be on "/sys/unit"
    And I should see a success notification "Unit organisasi ditambahkan."
    And I should see unit "E2E Pengurus Wilayah Satu" with kind "Teritorial" in the list

  Scenario: Create child unit under parent
    Given I am logged in as an admin
    And parish "E2E Paroki Unit Hierarki" exists
    And unit "E2E Unit Induk" exists in parish "E2E Paroki Unit Hierarki" with kind "council"
    And I open "/sys/unit"
    When I click the add unit button
    Then I should be on "/sys/unit/create"
    When I fill the unit name with "E2E Seksi Liturgi Anak"
    And I fill the unit code with "SL-01"
    And I select unit kind "Seksi" in unit form
    And I select parish "E2E Paroki Unit Hierarki" in unit form
    And I select parent unit "E2E Unit Induk" in unit form
    And I submit the unit form
    Then I should be on "/sys/unit"
    And I should see a success notification "Unit organisasi ditambahkan."
    And I should see unit "E2E Seksi Liturgi Anak" with kind "Seksi" in the list

  Scenario: Unit form validation
    Given I am logged in as an admin
    And I open "/sys/unit"
    When I click the add unit button
    Then I should be on "/sys/unit/create"
    When I submit the unit form
    Then I should see a unit validation error "Nama unit wajib diisi."
    And I should see a unit validation error "Jenis unit wajib dipilih."
    And I should see a unit validation error "Paroki wajib dipilih."
    When I select unit kind "Teritorial" in unit form
    And I submit the unit form
    Then I should see a unit validation error "Wilayah teritorial wajib dipilih."
    When I cancel the unit form
    Then I should be on "/sys/unit"

  Scenario: Edit an existing unit
    Given I am logged in as an admin
    And parish "E2E Paroki Unit Edit" exists
    And unit "E2E Unit To Edit" exists in parish "E2E Paroki Unit Edit" with kind "category"
    And I open "/sys/unit"
    When I click edit for unit "E2E Unit To Edit"
    And I fill the unit name with "E2E Unit Edited"
    And I fill the unit code with "CAT-EDITED"
    And I submit the unit form
    Then I should be on "/sys/unit"
    And I should see a success notification "Unit organisasi diperbarui."
    And I should see unit "E2E Unit Edited" in the list
    And I should not see unit "E2E Unit To Edit" in the list

  Scenario: Delete a unit
    Given I am logged in as an admin
    And parish "E2E Paroki Unit Delete" exists
    And unit "E2E Unit To Delete" exists in parish "E2E Paroki Unit Delete" with kind "committee"
    And I open "/sys/unit"
    When I click delete for unit "E2E Unit To Delete"
    And I confirm the deletion
    Then I should see a success notification "Unit organisasi dihapus."
    And I should not see unit "E2E Unit To Delete" in the list

  Scenario: Search unit by name
    Given I am logged in as an admin
    And parish "E2E Paroki Unit Search" exists
    And unit "E2E Unit Alfa" exists in parish "E2E Paroki Unit Search" with kind "section"
    And unit "E2E Unit Beta" exists in parish "E2E Paroki Unit Search" with kind "section"
    And I open "/sys/unit"
    When I search unit for "Alfa"
    Then I should see unit "E2E Unit Alfa" in the list
    And I should not see unit "E2E Unit Beta" in the list

