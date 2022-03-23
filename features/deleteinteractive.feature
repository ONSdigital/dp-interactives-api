Feature: Interactives API (Delete interactive)

    Scenario: Delete failed if interactive not in DB
        Given I am an interactives user
        When I DELETE "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
        Then the HTTP status code should be "404"

    Scenario: Successful delete
        Given I am an interactives user
        And I have these interactives:
                """
                [
                    {
                        "active": true,
                        "metadata": {
                            "label": "ad fugiat cillum",
                            "internal_id": ""
                        },
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When I DELETE "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
        Then the HTTP status code should be "200"