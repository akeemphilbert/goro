Feature: Binary File Storage
  As a pod user
  I want to store binary files like images and documents
  So that I can manage all my data in one place

  Background:
    Given the storage system is running
    And the pod storage is available

  Scenario: Store and retrieve a binary image file
    Given I have a binary image file "test.jpg"
    When I upload the file with content type "image/jpeg"
    Then the file should be stored successfully
    And I should be able to retrieve the exact original content
    And the MIME type should be preserved as "image/jpeg"

  Scenario: Store and retrieve a binary document file
    Given I have a binary document file "document.pdf"
    When I upload the file with content type "application/pdf"
    Then the file should be stored successfully
    And I should be able to retrieve the exact original content
    And the MIME type should be preserved as "application/pdf"

  Scenario: Store large binary file with streaming
    Given I have a large binary file of 10MB
    When I upload the file using streaming
    Then the file should be stored successfully
    And I should be able to retrieve it using streaming
    And the content should match exactly

  Scenario: Preserve file metadata
    Given I have a binary file with custom metadata
    When I upload the file with additional headers
    Then the file should be stored successfully
    And all metadata should be preserved
    And I should be able to retrieve the metadata

  Scenario: Handle binary file storage failure
    Given the storage system has limited space
    When I try to upload a file that exceeds available space
    Then I should receive a 507 Insufficient Storage response
    And the error message should indicate storage limitation

  Scenario: Verify binary file integrity
    Given I have uploaded a binary file
    When I retrieve the file
    Then the checksum should match the original
    And the file size should be identical
    And no data corruption should be detected