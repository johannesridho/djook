var express = require('express')
var bodyParser = require('body-parser')

// setup
var port = process.env.PORT || 3000
var app = express()

app.use(bodyParser.urlencoded({ limit: '50mb', extended: false }))
app.use(bodyParser.json({ limit: '50mb' }))


// routes
app.get('/', function (req, res) {
  res.send(JSON.stringify({ Hi: 'from djook' }))
});

app.post('/search-person', function (req, res) {
  var postImages = req.body.image

  var searchingPerson = require('./searchPerson')

  searchingPerson(postImages).then(function(result) {
    res.send(JSON.stringify(result))
  })

})

// start
app.listen(port, function () {
  console.log(`Example app listening on port !`);
});
