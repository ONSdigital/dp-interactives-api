Feature: Interactives API (publish interactive)
    Scenario: Publishing fails if not in correct state
        Given I have these interactives:
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
        When As an interactives user I PUT no file with form-data "/v1/collection/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "published": true,
                    "metadata": {
                        "collection_id": "col123",
                        "label": "Title321",
                        "title": "Title123",
                        "slug": "Title321",
                        "resource_id": "resid321",
                        "internal_id": "123"
                    }
                }
            """
        Then the HTTP status code should be "405"