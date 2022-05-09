Feature: Interactives API (Get interactive)

    Scenario: POST return unauthorized
        When I POST "/v1/interactives"
        """
            {
            }
        """
        Then the HTTP status code should be "403"

    Scenario: Update failed if validation rules not followed
        When As an interactives user I POST file "resources/interactives.zip" with form-data "/v1/interactives"
            """
                {
                    "metadata": { }
                }
            """
        Then the HTTP status code should be "400"
        And I should receive the following JSON response:
            """
                {
                    "errors": [
                        "Interactive.Metadata.Title",
                        "Interactive.Metadata.Label",
                        "Interactive.Metadata.InternalID"
                    ]
                }
            """
        
    Scenario: New interactive success with file
        When As an interactives user I POST file "resources/interactives.zip" with form-data "/v1/interactives"
            """
                {
                    "metadata": {
                        "label": "Title123",
                        "slug": "Title123",
                        "title": "Title123",
                        "resource_id": "AbcdE123",
                        "internal_id": "123"
                    }
                }
            """
        Then I should receive the following model response with status "202":
            """
                {
                    "id": "00000000-0000-0000-0000-000000000000",
                    "published": false,
                    "archive": {
                        "name":"rhyCq4GCknxx0nzeqx2LE077Ruo=/interactives.zip"
                    },
                    "metadata": {
                        "label": "Title123",
                        "slug": "Title123",
                        "title": "Title123",
                        "resource_id": "AbcdE123",
                        "internal_id": "123"
                    },
                    "state": "ArchiveUploaded",
                    "url": "http://localhost:27300/interactives/Title123-AbcdE123/embed"
                }
            """