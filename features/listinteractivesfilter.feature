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
                        "title": "title123",
                        "label": "Title123",
                        "internal_id": "123"
                    },
                    "state": "ImportSuccess"
                },
                {
                    "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                    "active": true,
                    "file_name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip",
                    "published" : false,
                    "metadata": {
                        "title": "title123",
                        "label": "Title321",
                        "internal_id": "123"
                    },
                    "state": "ImportSuccess"
                }
            ]
            """
        When As an interactives user with filter I GET '/v1/interactives?filter=%7B%22label%22%3A%20%20%22Title321%22%7D'
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
                                "title": "title123",
                                "label": "Title321",
                                "internal_id": "123"
                            },
                            "state": "ImportSuccess",
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

        Scenario: GET interactives (given a resource_id)
        Given I have these interactives:
            """
            [
                {
                    "id": "671375fa-2fc4-41cc-b845-ad04a56d96a7",
                    "active": true,
                    "published" : false,
                    "metadata": {
                        "title": "title123",
                        "label": "Title123",
                        "resource_id": "resid123",
                        "internal_id": "123"
                    },
                    "state": "ImportSuccess"
                },
                {
                    "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                    "active": true,
                    "file_name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip",
                    "published" : false,
                    "metadata": {
                        "title": "title123",
                        "label": "Title123",
                        "resource_id": "resid321",
                        "internal_id": "123"
                    },
                    "state": "ImportSuccess"
                }
            ]
            """
        When As an interactives user with filter I GET '/v1/interactives?filter=%7B%22resource_id%22%3A%20%20%22resid321%22%7D'
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
                                "title": "title123",
                                "label": "Title123",
                                "resource_id": "resid321",
                                "internal_id": "123"
                            },
                            "state": "ImportSuccess",
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