Feature: Interactives API (List interactives) - from public web access

    Scenario: GET a list of all interactives (skip deleted and unpublished)
        And I have these interactives:
            """
            [
                {
                    "id": "671375fa-2fc4-41cc-b845-ad04a56d0001",
                    "active": false,
                    "published" : false,
                    "metadata": {
                        "title": "title123",
                        "label": "ad fugiat cillum",
                        "internal_id": "123",
                        "resource_id": "abcde123",
                        "slug": "slug"
                    },
                    "state": "ImportSuccess"
                },
                {
                    "id": "671375fa-2fc4-41cc-b845-ad04a56d0002",
                    "active": true,
                    "file_name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip",
                    "published" : false,
                    "metadata": {
                        "title": "title123",
                        "label": "ad fugiat cillum",
                        "internal_id": "123",
                        "resource_id": "abcde123",
                        "slug": "slug"
                    },
                    "state": "ImportSuccess"
                },
                {
                    "id": "671375fa-2fc4-41cc-b845-ad04a56d0003",
                    "active": true,
                    "file_name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip",
                    "published" : true,
                    "metadata": {
                        "title": "publishedTitle",
                        "label": "ad fugiat cillum",
                        "internal_id": "456",
                        "resource_id": "abcde123",
                        "slug": "slug"
                    },
                    "state": "ImportSuccess"
                }
            ]
            """
        When I GET "/v1/interactives"
        Then I should receive the following list(model) response with status "200":
            """
                [
                        {
                            "id": "671375fa-2fc4-41cc-b845-ad04a56d0003",
                            "published" : true,
                            "archive": {
                               "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "metadata": {
                                "title": "publishedTitle",
                                "label": "ad fugiat cillum",
                                "internal_id": "456",
                                "resource_id": "abcde123",
                                "slug": "slug"
                            },
                            "state": "ImportSuccess",
                            "last_updated":"2021-01-01T00:00:02Z",
                            "url": "http://localhost:27300/interactives/slug-abcde123/embed"
                        }
                ]
            """