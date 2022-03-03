Feature: Interactives API (Delete interactive)

    Scenario: Delete failed if interactive not in DB
        When I DELETE "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
        Then the HTTP status code should be "404"

    Scenario: Successful delete
    Given I have these interactives:
            """
            [
                {
                    "_id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
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
                }
            ]
            """
    When I DELETE "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
    Then the HTTP status code should be "200"