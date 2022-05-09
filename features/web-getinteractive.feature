Feature: Interactives API (Get interactive) - from public web access

    Scenario: Unpublished public access
        And I have these interactives:
            """
            [
                {
                    "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                    "active": true,
                    "published": false,
                    "archive": {
                        "name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "last_updated": "2022-03-02T16:44:32.443Z",
                    "metadata": {
                        "title": "title123",
                        "label": "ad fugiat cillum",
                        "internal_id": "123"
                    },
                    "sha": "rhyCq4GCknxx0nzeqx2LE077Ruo=",
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I GET "/v1/interactives/ca99d09c-953a-4fe5-9b0a-51b3d40c01f7"
        Then the HTTP status code should be "404"

    Scenario: Published public access
        And I have these interactives:
            """
            [
                {
                    "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                    "active": true,
                    "published": true,
                    "archive": {
                        "name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "last_updated": "2022-03-02T16:44:32.443Z",
                    "metadata": {
                        "title": "title123",
                        "label": "ad fugiat cillum",
                        "internal_id": "123",
                        "resource_id": "abcde123",
                        "slug": "slug"
                    },
                    "sha": "rhyCq4GCknxx0nzeqx2LE077Ruo=",
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I GET "/v1/interactives/ca99d09c-953a-4fe5-9b0a-51b3d40c01f7"
        Then I should receive the following model response with status "200":
            """
                {
                    "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                    "published": true,
                    "archive": {
                        "name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "metadata": {
                        "title": "title123",
                        "label": "ad fugiat cillum",
                        "internal_id": "123",
                        "resource_id": "abcde123",
                        "slug": "slug"
                    },
                    "state": "ArchiveUploaded",
                    "last_updated":"2021-01-01T00:00:00Z",
                    "url": "http://localhost:27300/interactives/slug-abcde123/embed"
                }
            """