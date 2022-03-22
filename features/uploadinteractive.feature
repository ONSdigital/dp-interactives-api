Feature: Interactives API (Get interactive)

    Scenario: POST return unauthorized
        When I POST "/v1/interactives"
        """
            {
            }
        """
        Then the HTTP status code should be "403"
        
    Scenario: New interactive success with file
        When As an interactives user I POST file "resources/interactives.zip" with form-data "/v1/interactives"
            """
                {
                    "interactive": {
                        "metadata": {
                            "title": "ad fugiat cillum",
                            "primary_topic": "primary topic",
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
                                "keyword1"
                            ],
                            "meta_description": "cillum Excepteur",
                            "source": "reprehenderit do",
                            "summary": "aliqua Ut amet laboris exercitation"
                        }
                    }
                }
            """
        Then I should receive the following model response with status "202":
            """
                {
                    "id": "00000000-0000-0000-0000-000000000000",
                    "published": false,
                    "archive": {
                        "name":"rhyCq4GCknxx0nzeqx2LE077Ruo=/interactives.zip"
                    },
                    "metadata": {
                        "resource_id": "AbcdE123",
                        "title": "ad fugiat cillum",
                        "primary_topic": "primary topic",
                        "slug": "ad fugiat cillum",
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
                            "keyword1"
                        ],
                        "meta_description": "cillum Excepteur",
                        "source": "reprehenderit do",
                        "summary": "aliqua Ut amet laboris exercitation"
                    }
                }
            """