Feature: Interactives API (List interactives)

    Scenario: GET a list of all interactives (skip deleted)
        Given I have these interactives:
            """
            [
                {
                    "_id": "671375fa-2fc4-41cc-b845-ad04a56d96a7",
                    "file_name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip",
                    "last_updated": {
                        "$date": "2022-02-15T18:57:16.359Z"
                    },
                    "metadata": "{\"metadata1\":\"value1\",\"metadata2\":\"value2\",\"metadata3\":\"value3\"}",
                    "sha": "rhyCq4GCknxx0nzeqx2LE077Ruo=",
                    "state": "IsDeleted"
                },
                {
                    "_id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                    "file_name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip",
                    "last_updated": {
                        "$date": "2022-02-15T19:00:46.816Z"
                    },
                    "metadata": "{\"metadata1\":\"val1\",\"metadata2\":\"val2\",\"metadata3\":\"val3\",\"metadata5\":\"val5\"}",
                    "sha": "kqA7qPo1GeOJeff69lByWLbPiZM=",
                    "state": "ImportSuccess"
                }
            ]
            """
        When I GET "/interactives?limit=10&offset=0"
        Then I should receive the following JSON response with status "200":
            """
                {
                    "items": [
                        {
                            "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                            "metadata": {
                                "metadata1": "val1",
                                "metadata2": "val2",
                                "metadata3": "val3",
                                "metadata5": "val5"
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
        When I GET "/interactives?limit=10&offset=0"
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