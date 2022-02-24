Feature: Interactives API (Update interactive)

    Scenario: Update failed if no message body
        When I PUT "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "baddata": true
                }
            """
        Then the HTTP status code should be "400"

    Scenario: Update failed if interactive not in DB
        When I PUT "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "importstatus": true,
                    "metadata": {"metadata1" : "updatedval1", "metadata5" : "val5"}
                }
            """
        Then the HTTP status code should be "404"

    Scenario: Update failed if interactive is deleted
        Given I have these interactives:
                """
                [
                    {
                        "_id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                        "file_name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip",
                        "last_updated": {
                            "$date": "2022-02-08T19:04:52.891Z"
                        },
                        "metadata": "{\"metadata1\":\"XXX\",\"metadata2\":\"YYY\",\"metadata3\":\"ZZZ\"}",
                        "sha": "kqA7qPo1GeOJeff69lByWLbPiZM=",
                        "state": "IsDeleted"
                    }
                ]
                """
        When I PUT "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "importstatus": true,
                    "metadata": {"metadata1" : "updatedval1", "metadata5" : "val5"}
                }
            """
        Then the HTTP status code should be "404"

    Scenario: Update success
        Given I have these interactives:
                """
                [
                    {
                        "_id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                        "file_name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip",
                        "last_updated": {
                            "$date": "2022-02-08T19:04:52.891Z"
                        },
                        "metadata": "{\"metadata1\":\"value1\",\"metadata2\":\"value2\",\"metadata3\":\"value3\"}",
                        "sha": "kqA7qPo1GeOJeff69lByWLbPiZM=",
                        "state": "ArchiveUploaded"
                    }
                ]
                """
        When I PUT "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
            """
                {
                    "importstatus": true,
                    "metadata": {"metadata1" : "updatedvalue1", "metadata5" : "value5"}
                }
            """
        Then I should receive the following JSON response with status "200":
            """
                {
                    "metadata1": "updatedvalue1",
                    "metadata2": "value2",
                    "metadata3": "value3",
                    "metadata5": "value5"
                }
            """