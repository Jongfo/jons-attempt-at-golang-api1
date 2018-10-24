# jons-attempt-at-golang-api1
Assignment 1 for imt2681-2018


Dependencies:
https://github.com/marni/goigc
https://github.com/gorilla/mux



Usage:
goicd-jon.herokuapp.com/igcinfo/api
Returns meta data about api.
{
"uptime": <uptime>
"info": "Service for IGC tracks."
"version": "v1"
}


goicd-jon.herokuapp.com/igcinfo/api/igc
GET: returns json struct of track IDs
[<id1>, <id2>, ...]
  
POST: By posting a json with a url to a igc file we will regiser a track and return the ID
{
  "url": "<url>"
}

returns:
{
  "id": "<id>"
}


goicd-jon.herokuapp.com/igcinfo/api/igc/{ID}
returns json sturct with data on track by given ID. All are strings exept track_legth, which is a float64
{
"H_date": <date from File Header, H-record>,
"pilot": <pilot>,
"glider": <glider>,
"glider_id": <glider_id>,
"track_length": <calculated total track length>
}


goicd-jon.herokuapp.com/igcinfo/api/igc/{ID}/{field}
Returns plain text. Body will be empty if we didn't find what you're looking for.
The {field} is the data name as seen above. 
Exaple: "H_date" as {field} will return a text like "2016-02-19T00:00:00Z"
