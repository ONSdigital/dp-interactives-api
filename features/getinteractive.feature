Feature: Interactives API (Get interactive)

    Scenario: GET a specific interactives
        Given I have these interactives:
            """
            [
                {
                    "_id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                    "active": true,
                    "archive": {
                        "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                    },
                    "last_updated": "2022-03-02T16:44:32.443Z",
                    "metadata": {
                        "title": "ad fugiat cillum",
                        "primary_topic": "",
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
        When I GET "/v1/interactives/ca99d09c-953a-4fe5-9b0a-51b3d40c01f7"
        Then I should receive the following JSON response with status "200":
            """
                {
                    "title": "ad fugiat cillum",
                    "primary_topic": "",
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
                }
            """

    Scenario: GET a non-existing interactives
        Given I have these interactives:
            """
            [
                {
                    "_id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "active": true,
                    "file_name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip",
                    "last_updated": {
                        "$date": "2022-02-08T19:04:52.891Z"
                    },
                    "metadata": {
                        "title": "ad fugiat cillum",
                        "primary_topic": "",
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
                    "sha": "kqA7qPo1GeOJeff69lByWLbPiZM=",
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I GET "/v1/interactives/12345678-abb2-4432-ad22-9c23cf7ee222"
        Then the HTTP status code should be "404"