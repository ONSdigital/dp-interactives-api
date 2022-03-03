Feature: Interactives API (Get interactive)

    Scenario: GET a specific interactives
        Given I have these interactives:
            """
            [
                {
                    "active": true,
                    "metadata": "{\"metadata1\":\"XXX\",\"metadata2\":\"YYY\",\"metadata3\":\"ZZZ\"}",
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I GET "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
        Then I should receive the following JSON response with status "200":
            """
                {
                    "metadata1": "XXX",
                    "metadata2": "YYY",
                    "metadata3": "ZZZ"
                }
            """

    Scenario: GET a non-existing interactives
        Given I have these interactives:
            """
            [
                {
                    "active": true,
                    "metadata": "{\"metadata1\":\"XXX\",\"metadata2\":\"YYY\",\"metadata3\":\"ZZZ\"}",
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I GET "/v1/interactives/12345678-abb2-4432-ad22-9c23cf7ee222"
        Then the HTTP status code should be "404"