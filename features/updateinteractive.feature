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
                        "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                        "metadata": {
                            "label": "Title123",
                            "title": "Title123",
                            "resource_id": "resid321",
                            "internal_id": "123"
                        },
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When As an interactives user I PUT file "resources/interactives.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "interactive": {
                        "metadata": {
                            "label": "Title12345",
                            "title": "Title123",
                            "resource_id": "resid321",
                            "internal_id": "1234"
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
                            "label": "Title123",
                            "title": "Title123",
                            "slug": "Title123",
                            "resource_id": "resid321",
                            "internal_id": "123"
                        },
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When As an interactives user I PUT file "resources/interactives.zip" with form-data "/v1/interactives/ca99d09c-953a-4fe5-9b0a-51b3d40c01f7"
            """
                {
                    "interactive": {
                        "metadata": {
                            "label": "Title123",
                            "title": "Title123",
                            "slug": "Title321-update",
                            "resource_id": "resid321",
                            "internal_id": "123"
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
                            "title": "Title123",
                            "label": "Title123",
                            "slug": "Title123",
                            "resource_id": "resid321",
                            "internal_id": "123"
                        },
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When As an interactives user I PUT file "resources/interactives.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "interactive": {
                        "archive": {
                            "name":"kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                        },
                        "metadata": {
                            "title": "Title123",
                            "label": "Title123",
                            "slug": "Title123",
                            "resource_id": "resid321",
                            "internal_id": "123"
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
                        "title": "Title123",
                        "label": "Title123",
                        "slug": "Title123",
                        "resource_id": "resid321",
                        "internal_id": "123"
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
                            "label": "Title123",
                            "title": "Title123",
                            "slug": "human readable slug",
                            "resource_id": "resid321",
                            "internal_id": "123"
                        },
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When As an interactives user I PUT no file with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "interactive": {
                        "metadata": {
                            "label": "Title321",
                            "title": "Title123",
                            "slug": "Title321",
                            "resource_id": "resid321",
                            "internal_id": "123"
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
                        "label": "Title321",
                        "slug": "Title321",
                        "title": "Title123",
                        "resource_id": "resid321",
                        "internal_id": "123"
                    }
                }
            """