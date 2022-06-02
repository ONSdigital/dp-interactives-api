Feature: Interactives API (Patch interactive)

    Scenario: Non-existent interactive
        Given I am an interactives user
        When I PATCH "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "request_body": "doesnt matter here"
                }
            """
        Then the HTTP status code should be "404"

    Scenario: Bad patch request: invalid action
        Given I am an interactives user
        And I have these interactives:
            """
            [
                {
                    "active": true,
                    "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "metadata": {
                        "label": "Title123",
                        "title": "Title123",
                        "resource_id": "resid321",
                        "internal_id": "123"
                    },
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I PATCH "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "request_body": "does matter here"
                }
            """
        Then the HTTP status code should be "400"

    Scenario: Update success with new file
        Given I am an interactives user
        And I have these interactives:
            """
                [
                    {
                        "active": true,
                        "metadata": {
                            "title": "Title123",
                            "label": "Title123",
                            "slug": "Title123",
                            "resource_id": "resid321",
                            "internal_id": "123",
                            "collection_id": "a_collection"
                        },
                        "state": "ArchiveUploaded",
                        "last_updated":"2021-01-01T00:00:00Z"
                    }
                ]
            """
        When I PATCH "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "attribute": "Archive",
                    "interactive": {
                        "archive": {
                            "import_successful": true,
                            "import_message": "message",
                            "name": "f5XNzqLK76cMwldF835lkCuKO34=/single-interactive.zip",
                            "size_in_bytes": 86159,
                            "files": [
                                {
                                    "name": "interactives/15c0ae7d-2fc2-4adb-b6e5-f2f110f645d8/version-1/single-interactive/index.html",
                                    "mimetype": "text/html; charset=utf-8",
                                    "size_in_bytes": 47767
                                }
                            ]
                        }
                    }
                }
            """
        Then I should receive the following model response with status "200":
            """
                {
                    "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "published": false,
                    "archive": {
                        "import_message": "message",
                        "name": "f5XNzqLK76cMwldF835lkCuKO34=/single-interactive.zip",
                        "size_in_bytes": 86159,
                        "files": [
                            {
                                "name": "interactives/15c0ae7d-2fc2-4adb-b6e5-f2f110f645d8/version-1/single-interactive/index.html",
                                "mimetype": "text/html; charset=utf-8",
                                "size_in_bytes": 47767
                            }
                        ]
                    },
                    "metadata": {
                        "title": "Title123",
                        "label": "Title123",
                        "slug": "Title123",
                        "resource_id": "resid321",
                        "internal_id": "123",
                        "collection_id": "a_collection"
                    },
                    "state": "ImportSuccess",
                    "url": "http://localhost:27300/interactives/Title123-resid321/embed",
                    "uri": "/interactives/Title123-resid321"
                }
            """