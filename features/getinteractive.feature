Feature: Interactives API (Get interactive)

    Scenario: GET a specific interactive
        Given I am an interactives user
        And I have these interactives:
            """
            [
                {
                    "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                    "active": true,
                    "published": false,
                    "archive": {
                        "name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "last_updated": "2022-03-02T16:44:32.443Z",
                    "metadata": {
                        "title": "title123",
                        "label": "ad fugiat cillum",
                        "internal_id": "123"
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
                    "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                    "published": false,
                    "archive": {
                        "name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "metadata": {
                        "title": "title123",
                        "label": "ad fugiat cillum",
                        "internal_id": "123"
                    },
                    "state": "ArchiveUploaded",
                    "last_updated":"2021-01-01T00:00:00Z"
                }
            """

    Scenario: GET a non-existing interactives
        Given I am an interactives user
        And I have these interactives:
            """
            [
                {
                    "active": true,
                    "published": false,
                    "metadata": {
                        "title": "title123",
                        "label": "ad fugiat cillum",
                        "internal_id": "123"
                    },
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I GET "/v1/interactives/12345678-abb2-4432-ad22-9c23cf7ee222"
        Then the HTTP status code should be "404"

    Scenario: Unauthorised user access 
        Given I have these interactives:
            """
            [
                {
                    "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                    "active": true,
                    "published": false,
                    "archive": {
                        "name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "last_updated": "2022-03-02T16:44:32.443Z",
                    "metadata": {
                        "title": "title123",
                        "label": "ad fugiat cillum",
                        "internal_id": "123"
                    },
                    "sha": "rhyCq4GCknxx0nzeqx2LE077Ruo=",
                    "state": "ArchiveUploaded"
                }
            ]
            """
        When I GET "/v1/interactives/ca99d09c-953a-4fe5-9b0a-51b3d40c01f7"
        Then the HTTP status code should be "403"