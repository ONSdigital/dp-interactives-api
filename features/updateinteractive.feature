Feature: Interactives API (Update interactive)

    Scenario: Update failed if no message body
        When I PUT "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "baddata": true
                }
            """
        Then the HTTP status code should be "400"

    Scenario: Update failed if interactive not in DB
        When I PUT "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "import_successful": true,
                    "interactive": {
                        "metadata": {"metadata1" : "updatedval1", "metadata5" : "val5"}
                    }
                }
            """
        Then the HTTP status code should be "404"

    Scenario: Update failed if interactive is deleted
        Given I have these interactives:
                """
                [
                    {
                        "_id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                        "active": false,
                        "file_name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip",
                        "last_updated": {
                            "$date": "2022-02-08T19:04:52.891Z"
                        },
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
                        "sha": "kqA7qPo1GeOJeff69lByWLbPiZM=",
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When I PUT "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "import_successful": true,
                    "interactive": {
                        "metadata": {
                            "title": "ad fugiat cillum",
                            "primary_topic": "updated primary topic",
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
                    }
                }
            """
        Then the HTTP status code should be "404"

    Scenario: Update success
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
                        "sha": "kqA7qPo1GeOJeff69lByWLbPiZM=",
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When I PUT "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "import_successful": true,
                    "interactive": {
                        "metadata": {
                            "title": "ad fugiat cillum",
                            "primary_topic": "updated primary topic",
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
                    }
                }
            """
        Then I should receive the following JSON response with status "200":
            """
                {
                    "title": "ad fugiat cillum",
                    "primary_topic": "updated primary topic",
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