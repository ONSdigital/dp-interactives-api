Feature: Interactives API (List interactives)

    Scenario: GET a list of all interactives (skip deleted)
        Given I am an interactives user
        And I have these interactives:
            """
            [
                {
                    "id": "671375fa-2fc4-41cc-b845-ad04a56d96a7",
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
                    "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
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
                }
            ]
            """
        When I GET "/v1/interactives?limit=10&offset=0"
        Then I should receive the following list(model) response with status "200":
            """
                {
                    "items": [
                        {
                            "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                            "published" : false,
                            "archive": {
                               "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "metadata": {
                                "title": "title123",
                                "label": "ad fugiat cillum",
                                "internal_id": "123",
                                "resource_id": "abcde123",
                                "slug": "slug"
                            },
                            "state": "ImportSuccess",
                            "archive": {
                                "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "last_updated":"2021-01-01T00:00:01Z",
                            "url": "http://localhost:27300/interactives/slug-abcde123/embed"
                        }
                    ],
                    "count": 1,
                    "offset": 0,
                    "limit": 10,
                    "total_count": 1
                }
            """

    Scenario: GET returns an empty array if nothing in the database
        Given I am an interactives user
        And I have these interactives:
            """
            []
            """
        When I GET "/v1/interactives?limit=10&offset=0"
        Then I should receive the following list(model) response with status "200":
            """
                {
                    "items": [],
                    "count": 0,
                    "offset": 0,
                    "limit": 10,
                    "total_count": 0
                }
            """