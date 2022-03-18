Feature: Interactives API (Update interactive)

    Scenario: Update failed if no message body
        Given I am an interactives user
        When I PUT "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "baddata": true
                }
            """
        Then the HTTP status code should be "400"

    Scenario: Update failed if interactive not in DB
        When As an interactives user I PUT file "resources/interactives.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
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
                        "active": false,
                        "metadata": {
                            "title": "ad fugiat cillum",
                            "primary_topic": "updated primary topic",
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
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When As an interactives user I PUT file "resources/interactives.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "import_successful": true,
                    "interactive": {
                        "metadata": {
                            "title": "ad fugiat cillum",
                            "primary_topic": "updated primary topic",
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
                        }
                    }
                }
            """
        Then the HTTP status code should be "404"

    Scenario: Slug update for a published interactive is forbidden
        Given I have these interactives:
                """
                [
                    {
                        "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                        "active": true,
                        "published": true,
                        "metadata": {
                            "title": "ad fugiat cillum",
                            "slug": "human readable slug",
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
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When As an interactives user I PUT file "resources/interactives.zip" with form-data "/v1/interactives/ca99d09c-953a-4fe5-9b0a-51b3d40c01f7"
            """
                {
                    "import_successful": true,
                    "interactive": {
                        "metadata": {
                            "title": "ad fugiat cillum",
                            "primary_topic": "updated primary topic",
                            "slug": "a different human readable slug",
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
        Then the HTTP status code should be "403"

    Scenario: Update success with new file
        Given I have these interactives:
                """
                [
                    {
                        "active": true,
                        "metadata": {
                            "title": "ad fugiat cillum",
                            "slug": "human readable slug",
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
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When As an interactives user I PUT file "resources/interactives.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "import_successful": true,
                    "interactive": {
                        "archive": {
                            "name":"kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                        },
                        "metadata": {
                            "title": "ad fugiat cillum [should not get updated]",
                            "slug": "updated human readable slug",
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
                    "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "published": false,
                    "archive": {
                        "name":"kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "metadata": {
                        "title": "ad fugiat cillum",
                        "primary_topic": "updated primary topic",
                        "slug": "updated human readable slug",
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
            """

    Scenario: Update success without a new file
        Given I have these interactives:
                """
                [
                    {
                        "active": true,
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
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When As an interactives user I PUT no file with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "import_successful": true,
                    "interactive": {
                        "metadata": {
                            "title": "ad fugiat cillum [should not get updated]",
                            "primary_topic": "updated primary topic",
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
                        }
                    }
                }
            """
        Then I should receive the following JSON response with status "200":
            """
                {
                    "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "published": false,
                    "archive": {
                        "name":"kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "metadata": {
                        "title": "ad fugiat cillum",
                        "primary_topic": "updated primary topic",
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
                    }
                }
            """