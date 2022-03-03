Feature: Interactives API (List interactives)

    Scenario: GET a list of all interactives (skip deleted)
        Given I have these interactives:
            """
            [
                {
                    "_id": "bd459093-1207-45f6-981f-d12522ffc499",
                    "active": false,
                    "archive": {},
                    "last_updated": "2022-03-02T16:23:05.201Z",
                    "metadata": {
                        "title": "ad fugiat cillum12",
                        "primary_topic": "",
                        "topics": [
                        "topic1",
                        "topic2",
                        "topic3"
                        ],
                        "surveys": [
                        "survey1",
                        "survey2"
                        ],
                        "release_date": "0001-01-01T00:00:00.000Z",
                        "uri": "id occaecat do",
                        "edition": "in quis cupidatat tempor",
                        "keywords": [
                        "keywd1"
                        ],
                        "meta_description": "cillum Excepteur",
                        "source": "reprehenderit do",
                        "summary": "aliqua Ut amet laboris exercitation"
                    },
                    "sha": "PQ3EkWb2MQ0l5TLc9jZM8RiY2j0=",
                    "state": "ImportSuccess"
                    },
                    {
                    "_id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                    "active": true,
                    "archive": {
                        "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                    },
                    "last_updated": "2022-03-02T16:44:32.443Z",
                    "metadata": {
                        "title": "ad fugiat cillum",
                        "primary_topic": "commodo sint labore",
                        "topics": [
                        "topic1"
                        ],
                        "surveys": [
                        "survey1"
                        ],
                        "release_date": "2022-03-01T22:04:06.311Z",
                        "uri": "id occaecat do",
                        "edition": "in quis cupidatat tempor",
                        "keywords": [
                        "keywd1"
                        ],
                        "meta_description": "cillum Excepteur",
                        "source": "reprehenderit do",
                        "summary": "aliqua Ut amet laboris exercitation"
                    },
                    "sha": "rhyCq4GCknxx0nzeqx2LE077Ruo=",
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I GET "/v1/interactives?limit=10&offset=0"
        Then I should receive the following JSON response with status "200":
            """
                {
                    "items": [
                        {
                            "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                            "metadata": {
                                "title": "ad fugiat cillum",
                                "primary_topic": "commodo sint labore",
                                "topics": [
                                    "topic1"
                                ],
                                "surveys": [
                                    "survey1"
                                ],
                                "release_date": "2022-03-01T22:04:06.311Z",
                                "uri": "id occaecat do",
                                "edition": "in quis cupidatat tempor",
                                "keywords": [
                                    "keywd1"
                                ],
                                "meta_description": "cillum Excepteur",
                                "source": "reprehenderit do",
                                "summary": "aliqua Ut amet laboris exercitation"
                            },
                            "archive": {
                                "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            }
                        }
                    ],
                    "count": 1,
                    "offset": 0,
                    "limit": 10,
                    "total_count": 1
                }
            """

    Scenario: GET returns an empty array if nothing in the database
        Given I have these interactives:
            """
            []
            """
        When I GET "/v1/interactives?limit=10&offset=0"
        Then I should receive the following JSON response with status "200":
            """
                {
                    "items": [],
                    "count": 0,
                    "offset": 0,
                    "limit": 10,
                    "total_count": 0
                }
            """