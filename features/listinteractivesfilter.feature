Feature: Interactives API (List interactives)

    Scenario: GET interactives (title - string)
        Given I have these interactives:
            """
            [
                {
                    "id": "671375fa-2fc4-41cc-b845-ad04a56d96a7",
                    "active": true,
                    "published" : false,
                    "metadata": {
                        "title": "Title123",
                        "collectionID" : "",
                        "primary_topic": "",
                        "slug": "human readable slug",
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
                    "state": "ImportSuccess"
                },
                {
                    "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                    "active": true,
                    "file_name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip",
                    "published" : false,
                    "metadata": {
                        "title": "Title321",
                        "collectionID" : "",
                        "primary_topic": "",
                        "slug": "human readable slug",
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
                    "state": "ImportSuccess"
                }
            ]
            """
        When As an interactives user with filter I GET '/v1/interactives?filter=%7B%22title%22%3A%20%20%22Title321%22%7D'
        Then I should receive the following JSON response with status "200":
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
                                "title": "Title321",
                                "collectionID" : "",
                                "primary_topic": "",
                                "slug": "human readable slug",
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
                    "limit": 20,
                    "total_count": 1
                }
            """

    Scenario: Scenario: GET interactives (topics - array)
        Given I have these interactives:
            """
            [
                {
                    "id": "671375fa-2fc4-41cc-b845-ad04a56d96a7",
                    "active": true,
                    "published" : false,
                    "metadata": {
                        "title": "Title123",
                        "collectionID" : "",
                        "primary_topic": "",
                        "slug": "human readable slug",
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
                    "state": "ImportSuccess"
                },
                {
                    "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                    "active": true,
                    "file_name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip",
                    "published" : false,
                    "metadata": {
                        "title": "Title321",
                        "collectionID" : "",
                        "primary_topic": "",
                        "slug": "human readable slug",
                        "topics": [
                        "topic2"
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
                    "state": "ImportSuccess"
                }
            ]
            """
        When As an interactives user with filter I GET '/v1/interactives?filter=%7B%22topics%22%3A%20%20%5B%22topic2%22%5D%7D'
        Then I should receive the following JSON response with status "200":
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
                                "title": "Title321",
                                "collectionID" : "",
                                "primary_topic": "",
                                "slug": "human readable slug",
                                "topics": [
                                "topic2"
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
                    "limit": 20,
                    "total_count": 1
                }
            """