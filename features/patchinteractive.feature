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

    Scenario: Update success with new archive
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
                            "size_in_bytes": 86159
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
                        "size_in_bytes": 86159
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
                    "url": "http://preview_url/interactives/Title123-resid321/embed",
                    "uri": "/interactives/Title123-resid321"
                }
            """

    Scenario: Update success with new archive file
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
                    "attribute": "ArchiveFile",
                    "archive_files": [
                        {
                            "name": "index.html",
                            "size_in_bytes": 86159,
                            "mimetype": "text/html",
                            "uri": "index.html"
                        }
                    ]
                }
            """
        Then I should receive the following model response with status "200":
            """
                {
                    "last_updated":"2021-01-01T00:00:00Z",
                    "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "published": false,
                    "archive": {
                        "name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip",
                        "size_in_bytes": 0,
                        "upload_root_directory": ""
                    },
                    "metadata": {
                        "title": "Title123",
                        "label": "Title123",
                        "slug": "Title123",
                        "resource_id": "resid321",
                        "internal_id": "123",
                        "collection_id": "a_collection"
                    },
                    "state": "ArchiveUploaded",
                    "url": "http://preview_url/interactives/Title123-resid321/embed",
                    "uri": "/interactives/Title123-resid321"
                }
            """
        And I should have these archive files:
            | InteractiveID                        | Name       | Mimetype  | Size  | URI        |
            | 0d77a889-abb2-4432-ad22-9c23cf7ee796 | index.html | text/html | 86159 | index.html |