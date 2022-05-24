Feature: Interactives API (publish interactive)
    Scenario: Publishing fails if not in correct state
        Given I am an interactives user
        And I have these interactives:
                """
                [
                    {
                        "active": true,
                        "metadata": {
                            "label": "Title123",
                            "collection_id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                            "title": "Title123",
                            "slug": "human readable slug",
                            "resource_id": "resid321",
                            "internal_id": "123"
                        },
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When I PATCH "/v1/collection/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                }
            """
        Then the HTTP status code should be "409"
