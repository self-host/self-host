# First interactions with the Self-hosting API

**Before venturing further**, make sure that you've completed all the necessary steps outlined in [Five to fifteen-minute deployment](https://github.com/self-host/self-host/blob/main/docs/test_deployment.md) walkthrough.

![Dangerous][dangerous-to-go-alone]

Throughout this guide, we will use the Swagger UI interface embedded in the Self-host API server. Go ahead and access it now. Authenticate with any user you like, but the user requires full access to the entire system. For simplicity's sake, use the root user if you are using your local development environment.

**You did remember to change the server URL and to add your credentials via the authorize button, right!?**


# Create your first time-series

Let's start with something simple. Let's create a time series to store a temperature value.

1. Scroll down to the `timeseries` section and expand the green POST /v2/timeseries field.
2. Click on the *Try it out* button.
3. Start by erasing the `"thing_uuid": "e21ae595-15a5-4f11-8992-9d33600cc1ee",` line. There are no things in the system and indeed nothing with that specific UUID.
4. Then also erase the `tags: [...]` lines as we don't want any tags added to the time series at this point.
5. You should now have something that looks like this:  
```json
{
  "name": "My Time Series",
  "si_unit": "C",
  "lower_bound": -50,
  "upper_bound": 50
}
```
6. Click on the blue `Execute` button.
7. Scroll down below the response example. There should be a response looking similar to this, yet with another UUID.  
```json
{
  "created_by": "00000000-0000-1000-8000-000000000000",
  "lower_bound": -50,
  "name": "My Time Series",
  "si_unit": "C",
  "tags": [],
  "thing_uuid": null,
  "upper_bound": 50,
  "uuid": "6ee20313-67b7-4203-8c02-e882fe454fc3"
}
```
8. Take note of the UUID in the JSON response, as we will need it in the future.

![You received a new UUID!][you-received-an-item]


### Adding data to a time series

Now, let's add a few data points to our new time series.

1. Scroll down to the POST /v2/timeseries/{uuid}/data field and expand it.
2. Click on the *Try it out* button.
3. Erase the example UUID in the field and insert our new UUID.
4. Take note of the request body format. It's a list of objects where `v` stands for value and `ts` stands for the timestamp. The format for the timestamp is RFC-3339.
```json
[
  {
    "v": 3.14,
    "ts": "2021-06-22T13:20:55.286Z"
  }
]
```
5. Click on the blue `Execute` button.
6. Scroll down below the response example. As you can see, we got nothing more than an HTTP 201 (Created) response from the server. This response means that everything is OK and that the server stored the data in the database.
7. Let's scroll back up and click on the blue `Execute` button again.
8. The response we got this time is not 201 any longer, but 400 with the text: "*Request caused an error due to duplicate key violation*". This error means that there already exists a data point at that particular point in time for this time series. Two data points can occupy the same point in time. If you want to overwrite a data point, you first need to erase it.

### Retrieve data from a time series

### Delete data from a time series


# Creating a Thing

## Connect a time series to a Thing


# Create a dataset

## Updating the dataset

## Retrieving the content of a dataset


# Managing tags

## For a time series

## For a thing

## For a dataset

## For a program


# Finding stuff

## List time series by tags

## List things by tags

## List dataset by tags

## List programs by tags


# Create a Program

## Using selfctl to aid with software development

## Adding code to a program

## Signing the code of a program

## Activating the program

## Modifying a program


[dangerous-to-go-alone]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/its_dangerous_to_go_alone.png "It's dangerous to go alone"
[you-received-an-item]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/you_received_an_item.png "You received an item"