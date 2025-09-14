Feature: Container Performance and Large Collections
  As a pod user
  I want efficient container operations
  So that browsing large collections is fast

  Background:
    Given a clean LDP server is running
    And the server supports container operations

  Scenario: Large container pagination
    Given a container "large-collection" exists
    And the container has 1000 resources
    When I request the first page with 50 items
    Then the response should contain 50 items
    And the response should include pagination links
    And the response time should be under 1 second

  Scenario: Container member filtering
    Given a container "mixed-content" exists
    And the container has 100 RDF resources
    And the container has 100 binary files
    When I filter for RDF resources only
    Then the response should contain only RDF resources
    And the response should be returned quickly

  Scenario: Container member sorting
    Given a container "documents" exists
    And the container has resources with different creation dates
    When I sort members by creation date descending
    Then the members should be ordered by creation date
    And the newest resources should appear first

  Scenario: Streaming large container listings
    Given a container "huge-collection" exists
    And the container has 10000 resources
    When I request the container listing
    Then the response should be streamed
    And memory usage should remain constant
    And the response should start immediately

  Scenario: Concurrent container access
    Given a container "shared" exists
    When 10 clients simultaneously access the container
    Then all requests should succeed
    And the response times should be reasonable
    And no race conditions should occur

  Scenario: Container size caching
    Given a container "cached" exists
    And the container has 500 resources
    When I request container metadata multiple times
    Then the size information should be cached
    And subsequent requests should be faster
    And the cache should be invalidated on updates

  Scenario: Deep hierarchy navigation performance
    Given a container hierarchy 10 levels deep
    When I navigate to the deepest container
    Then the path resolution should be efficient
    And the response time should be acceptable
    And breadcrumb generation should be fast

  Scenario: Bulk operations performance
    Given a container "bulk-test" exists
    When I add 100 resources to the container in batch
    Then the operation should complete efficiently
    And the membership index should be updated correctly
    And memory usage should remain reasonable