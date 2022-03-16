Feature: Interactives API (Delete interactive)

    Scenario: Delete failed if interactive not in DB
        Given I am an interactives user
        When I DELETE "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
        Then the HTTP status code should be "404"

    Scenario: Successful delete
        Given I am an interactives user
        And I have these interactives:
                """
                [
                    {
                        "active": true,
                        "metadata": {
                            "title": "ad fugiat cillum",
                            "collectionID" : "",
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
        When I DELETE "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
        Then the HTTP status code should be "200"