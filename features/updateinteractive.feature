Feature: Interactives API (Update interactive)

    Scenario: Update failed if validation rules not followed
        When As an interactives user I PUT file "resources/single-interactive.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "metadata": { }
                }
            """
        Then the HTTP status code should be "400"
        And I should receive the following JSON response:
            """
                {
                    "errors": [
                        "interactive.metadata.title: required",
                        "interactive.metadata.label: required",
                        "interactive.metadata.internalid: required"
                    ]
                }
            """

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
        When As an interactives user I PUT file "resources/single-interactive.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "metadata": {
                        "label": "Title12345",
                        "title": "Title123",
                        "internal_id": "1234"
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
        When As an interactives user I PUT file "resources/single-interactive.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "metadata": {
                        "label": "Title12345",
                        "title": "Title123",
                        "resource_id": "resid321",
                        "internal_id": "1234"
                    }
                }
            """
        Then the HTTP status code should be "404"

    Scenario: Slug update for a published interactive is allowed - redirect logic means old url will still work
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
        When As an interactives user I PUT file "resources/single-interactive.zip" with form-data "/v1/interactives/ca99d09c-953a-4fe5-9b0a-51b3d40c01f7"
            """
                {
                    "metadata": {
                        "label": "Title456",
                        "title": "Title456",
                        "slug": "Title456",
                        "resource_id": "resid456",
                        "internal_id": "456"
                    }
                }
            """
        Then I should receive the following model response with status "200":
            """
                {
                    "id": "ca99d09c-953a-4fe5-9b0a-51b3d40c01f7",
                    "published": true,
                    "archive": {
                        "name": "single-interactive.zip",
                        "size_in_bytes": 591714,
                        "files": [
                            {
                                "name": "index.html",
                                "mimetype": "tbc",
                                "size_in_bytes": 47767,
                                "uri": "index.html"
                            }
                        ]
                    },
                    "html_files": [
                        {
                            "name": "index.html",
                            "uri": "/interactives/Title456-resid321/index.html"
                        }
                    ],
                    "metadata": {
                        "title": "Title456",
                        "label": "Title456",
                        "slug": "Title456",
                        "resource_id": "resid321",
                        "internal_id": "456"
                    },
                    "state": "ArchiveUploaded",
                    "url": "http://localhost:27300/interactives/Title456-resid321/embed",
                    "uri": "/interactives/Title456-resid321"
                }
            """

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
                        "state": "ArchiveUploaded",
                        "last_updated":"2021-01-01T00:00:00Z"
                    }
                ]
                """
        When As an interactives user I PUT file "resources/single-interactive.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "archive": {
                        "name":"kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "metadata": {
                        "title": "Title456",
                        "label": "Title456",
                        "slug": "Title456",
                        "internal_id": "456",
                        "resource_id": "should_not_update"
                    }
                }
            """
        Then I should receive the following model response with status "200":
            """
                {
                    "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "published": false,
                    "archive": {
                        "name": "single-interactive.zip",
                        "size_in_bytes": 591714,
                        "files": [
                            {
                                "name": "index.html",
                                "mimetype": "tbc",
                                "size_in_bytes": 47767,
                                "uri": "index.html"
                            }
                        ]
                    },
                    "html_files": [
                        {
                            "name": "index.html",
                            "uri": "/interactives/Title456-resid321/index.html"
                        }
                    ],
                    "metadata": {
                        "title": "Title456",
                        "label": "Title456",
                        "slug": "Title456",
                        "resource_id": "resid321",
                        "internal_id": "456"
                    },
                    "state": "ArchiveUploaded",
                    "url": "http://localhost:27300/interactives/Title456-resid321/embed",
                    "uri": "/interactives/Title456-resid321"
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
                    "metadata": {
                        "label": "Title321",
                        "title": "Title123",
                        "slug": "Title321",
                        "resource_id": "resid321",
                        "internal_id": "123"
                    }
                }
            """
        Then I should receive the following model response with status "200":
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
                    },
                    "state": "ArchiveUploaded",
                    "url": "http://localhost:27300/interactives/Title321-resid321/embed",
                    "uri": "/interactives/Title321-resid321"
                }
            """

    Scenario: Metadata update for a published interactive is allowed - redirect logic means old url will still work
        Given I have these interactives:
                """
                [
                    {
                        "active": true,
                        "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                        "published": true,
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
                    "metadata": {
                        "label": "Title456",
                        "title": "Title456",
                        "slug": "Title456",
                        "resource_id": "resid456",
                        "internal_id": "456"
                    }
                }
            """
        Then I should receive the following model response with status "200":
            """
                {
                    "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "published": true,
                    "archive": {
                        "name":"kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip"
                    },
                    "metadata": {
                        "label": "Title456",
                        "slug": "Title456",
                        "title": "Title456",
                        "resource_id": "resid321",
                        "internal_id": "456"
                    },
                    "state": "ArchiveUploaded",
                    "url": "http://localhost:27300/interactives/Title456-resid321/embed",
                    "uri": "/interactives/Title456-resid321"
                }
            """

    Scenario: Update success with only a new file for published interactive
        Given I have these interactives:
                """
                [
                    {
                        "published": true,
                        "active": true,
                        "metadata": {
                            "label": "Title123",
                            "title": "Title123",
                            "slug": "human readable slug",
                            "resource_id": "resid321",
                            "internal_id": "123"
                        },
                        "state": "ImportSuccess"
                    }
                ]
                """
        When As an interactives user I PUT file "resources/single-interactive.zip" with form-data "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {

                }
            """
        Then I should receive the following model response with status "200":
            """
                {
                    "id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "published": true,
                    "archive": {
                        "name": "single-interactive.zip",
                        "size_in_bytes": 591714,
                        "files": [
                            {
                                "name": "index.html",
                                "mimetype": "tbc",
                                "size_in_bytes": 47767,
                                "uri": "index.html"
                            }
                        ]
                    },
                    "html_files": [
                        {
                            "name": "index.html",
                            "uri": "/interactives/human readable slug-resid321/index.html"
                        }
                    ],
                    "metadata": {
                        "label": "Title123",
                        "slug": "human readable slug",
                        "title": "Title123",
                        "resource_id": "resid321",
                        "internal_id": "123"
                    },
                    "state": "ArchiveUploaded",
                    "url": "http://localhost:27300/interactives/human readable slug-resid321/embed",
                    "uri": "/interactives/human readable slug-resid321"
                }
            """