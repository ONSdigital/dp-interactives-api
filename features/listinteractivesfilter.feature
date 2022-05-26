Feature: Interactives API (List interactives)

    Scenario: GET interactives (filter by label)
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
                        "internal_id": "123",
                        "resource_id": "abcde1",
                        "slug": "slug"
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
                        "internal_id": "123",
                        "resource_id": "abcde2",
                        "slug": "slug"
                    },
                    "state": "ImportSuccess"
                }
            ]
            """
        When As an interactives user with filter I GET '/v1/interactives?filter=%7B%0A%20%20%22associate_collection%22%3A%20false%2C%0A%20%20%22metadata%22%3A%20%7B%0A%20%20%20%20%22label%22%3A%20%22Title321%22%0A%20%20%7D%0A%7D'
        Then I should receive the following list(model) response with status "200":
            """
                 [
                        {
                            "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                            "published" : false,
                            "archive": {
                               "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "metadata": {
                                "title": "title123",
                                "label": "Title321",
                                "internal_id": "123",
                                "resource_id": "abcde2",
                                "slug": "slug"
                            },
                            "state": "ImportSuccess",
                            "archive": {
                                "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "last_updated":"2021-01-01T00:00:01Z",
                            "url": "http://localhost:27300/interactives/slug-abcde2/embed",
                            "uri": "/interactives/slug-abcde2/embed"
                        }
                    ]
            """

        Scenario: GET interactives (filter by resource_id)
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
                        "resource_id": "resid1",
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
                        "resource_id": "resid2",
                        "internal_id": "123",
                        "slug": "slug"
                    },
                    "state": "ImportSuccess"
                }
            ]
            """
        When As an interactives user with filter I GET '/v1/interactives?filter=%7B%0A%20%20%22associate_collection%22%3A%20false%2C%0A%20%20%22metadata%22%3A%20%7B%0A%20%20%20%20%22resource_id%22%3A%20%22resid2%22%0A%20%20%7D%0A%7D'
        Then I should receive the following list(model) response with status "200":
            """
                [
                        {
                            "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                            "published" : false,
                            "archive": {
                               "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "metadata": {
                                "title": "title123",
                                "label": "Title123",
                                "resource_id": "resid2",
                                "internal_id": "123",
                                "slug": "slug"
                            },
                            "state": "ImportSuccess",
                            "archive": {
                                "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "last_updated":"2021-01-01T00:00:01Z",
                            "url": "http://localhost:27300/interactives/slug-resid2/embed",
                            "uri": "/interactives/slug-resid2/embed"
                        }
                    ]
            """

        Scenario: GET interactives (filter by associated collection-id - linked + exclude other linked)
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
                        "resource_id": "resid1",
                        "collection_id": "12345",
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
                        "resource_id": "resid2",
                        "internal_id": "123",
                        "collection_id": "54321",
                        "slug": "slug"
                    },
                    "state": "ImportSuccess"
                }
            ]
            """
        When As an interactives user with filter I GET '/v1/interactives?filter=%7B%0A%20%20%22associate_collection%22%3A%20true%2C%0A%20%20%22metadata%22%3A%20%7B%0A%20%20%20%20%22collection_id%22%3A%20%2254321%22%0A%20%20%7D%0A%7D'
        Then I should receive the following list(model) response with status "200":
            """
                [
                        {
                            "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                            "published" : false,
                            "archive": {
                               "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "metadata": {
                                "title": "title123",
                                "label": "Title123",
                                "resource_id": "resid2",
                                "internal_id": "123",
                                "collection_id": "54321",
                                "slug": "slug"
                            },
                            "state": "ImportSuccess",
                            "archive": {
                                "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "last_updated":"2021-01-01T00:00:01Z",
                            "url": "http://localhost:27300/interactives/slug-resid2/embed",
                            "uri": "/interactives/slug-resid2/embed"
                        }
                ]
            """

        Scenario: GET interactives (fliter by associated collection-id - linked + published)
        Given I have these interactives:
            """
            [
                {
                    "id": "671375fa-2fc4-41cc-b845-ad04a56d96a7",
                    "active": true,
                    "published" : true,
                    "metadata": {
                        "title": "title123",
                        "label": "Title123",
                        "resource_id": "resid1",
                        "internal_id": "123",
                        "slug": "slug"
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
                        "resource_id": "resid2",
                        "internal_id": "123",
                        "collection_id": "54321",
                        "slug": "slug"
                    },
                    "state": "ImportSuccess"
                }
            ]
            """
        When As an interactives user with filter I GET '/v1/interactives?filter=%7B%0A%20%20%22associate_collection%22%3A%20true%2C%0A%20%20%22metadata%22%3A%20%7B%0A%20%20%20%20%22collection_id%22%3A%20%2254321%22%0A%20%20%7D%0A%7D'
        Then I should receive the following list(model) response with status "200":
            """
                [
                    {
                            "id": "671375fa-2fc4-41cc-b845-ad04a56d96a7",
                            "published" : true,
                            "metadata": {
                                "title": "title123",
                                "label": "Title123",
                                "resource_id": "resid1",
                                "internal_id": "123",
                                "slug": "slug"
                            },
                            "state": "ImportSuccess",
                            "archive": {
                                "name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                            },
                            "last_updated":"2021-01-01T00:00:00Z",
                            "url": "http://localhost:27300/interactives/slug-resid1/embed",
                            "uri": "/interactives/slug-resid1/embed"
                        },
                        {
                            "id": "2683c698-e15b-4d32-a990-ba37d93a4d83",
                            "published" : false,
                            "metadata": {
                                "title": "title123",
                                "label": "Title123",
                                "resource_id": "resid2",
                                "internal_id": "123",
                                "collection_id": "54321",
                                "slug": "slug"
                            },
                            "state": "ImportSuccess",
                            "archive": {
                                "name": "rhyCq4GCknxx0nzeqx2LE077Ruo=/TestMe.zip"
                            },
                            "last_updated":"2021-01-01T00:00:01Z",
                            "url": "http://localhost:27300/interactives/slug-resid2/embed",
                            "uri": "/interactives/slug-resid2/embed"
                        }
                    ]
            """