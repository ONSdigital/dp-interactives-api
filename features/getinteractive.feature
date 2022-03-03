Feature: Interactives API (Get interactive)

    Scenario: GET a specific interactives
        Given I have these interactives:
            """
            [
                {
                    "active": true,
                    "metadata": "{\"metadata1\":\"XXX\",\"metadata2\":\"YYY\",\"metadata3\":\"ZZZ\"}",
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
                    "active": true,
                    "metadata": "{\"metadata1\":\"XXX\",\"metadata2\":\"YYY\",\"metadata3\":\"ZZZ\"}",
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I GET "/v1/interactives/12345678-abb2-4432-ad22-9c23cf7ee222"
        Then the HTTP status code should be "404"