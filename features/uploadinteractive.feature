Feature: Interactives API (Get interactive)

    Scenario: POST return unauthorized
        When I POST "/v1/interactives"
        """
            {
            }
        """
        Then the HTTP status code should be "401"

    Scenario: Update failed if validation rules not followed - missing mandatory attributes
        When As an interactives user I POST file "resources/single-interactive.zip" with form-data "/v1/interactives"
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
                        "interactive.metadata.title: required",
                        "interactive.metadata.label: required",
                        "interactive.metadata.internalid: required"
                    ]
                }
            """

    Scenario: Update failed if validation rules not followed - attributes not correct format
        When As an interactives user I POST file "resources/single-interactive.zip" with form-data "/v1/interactives"
            """
                {
                    "metadata": {
                        "title": " ",
                        "label": "only alphanum allowed",
                        "internal_id": "only alphanum allowed"
                    }
                }
            """
        Then the HTTP status code should be "400"
        And I should receive the following JSON response:
            """
                {
                    "errors": [
                        "interactive.metadata.title: required",
                        "interactive.metadata.label: alphanum",
                        "interactive.metadata.internalid: alphanum"
                    ]
                }
            """
        
    Scenario: New interactive success with file
        When As an interactives user I POST file "resources/single-interactive.zip" with form-data "/v1/interactives"
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
                        "name": "single-interactive.zip",
                        "size_in_bytes": 591714,
                        "files": [
                            {
                                "name": "index.html",
                                "size_in_bytes": 47767,
                                "uri": "index.html"
                            }
                        ]
                    },
                    "html_files": [
                        {
                            "name": "index.html",
                            "uri": "/interactives/Title123-AbcdE123/index.html"
                        }
                    ],
                    "metadata": {
                        "label": "Title123",
                        "slug": "Title123",
                        "title": "Title123",
                        "resource_id": "AbcdE123",
                        "internal_id": "123"
                    },
                    "state": "ArchiveUploaded",
                    "url": "http://localhost:27300/interactives/Title123-AbcdE123/embed",
                    "uri": "/interactives/Title123-AbcdE123"
                }
            """