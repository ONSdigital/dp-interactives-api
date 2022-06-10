Feature: Interactives API (Delete interactive)

    Scenario: Delete failed if interactive not in DB
        Given I am an interactives user
        When I DELETE "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
        Then the HTTP status code should be "404"

    Scenario: Cannot delete published interactive
        Given I am an interactives user
        And I have these interactives:
                """
                [
                    {
                        "active": true,
                        "published": true,
                        "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                        "metadata": {
                            "title": "title123",
                            "label": "ad fugiat cillum",
                            "internal_id": "123"
                        },
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When I DELETE "/v1/interactives/ca99d09c-953a-4fe5-9b0a-51b3d40c01f7"
        Then the HTTP status code should be "403"

    Scenario: Successful delete
        Given I am an interactives user
        And I have these interactives:
                """
                [
                    {
                        "active": true,
                        "published": false,
                        "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                        "metadata": {
                            "label": "ad fugiat cillum",
                            "internal_id": ""
                        },
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When I DELETE "/v1/interactives/ca99d09c-953a-4fe5-9b0a-51b3d40c01f7"
        Then the HTTP status code should be "204"