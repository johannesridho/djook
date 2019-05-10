var AWS = require('aws-sdk');
var uuid = require('uuid');

var persons = [
  {
    name: 'Angel',
    image: require('./dataset/angel.json')
  },
  {
    name: 'Putra',
    image: require('./dataset/putra.json')
  }
]

var awsConfig = {
  region: 'us-east-1'
}

if (process.env.AWS_ACCESS_KEY_ID) {
  awsConfig.accessKeyId = process.env.AWS_ACCESS_KEY_ID
  secretAccessKey = process.env.AWS_SECRET_ACCESS_KEY
  console.log('djaksjdkasjd ', process.env.AWS_ACCESS_KEY_ID)
}

var rekognition = new AWS.Rekognition(awsConfig);

var searchingPerson = function (name, params) {
  return new Promise((resolve) => {

    rekognition.compareFaces(params, function (err, data) {
      if (err) {

        throw new Error(err, err.stack);
        reject(err)

      } else {
        var result = {
          ...data,
          name
        }

        resolve(result)
      }
    });

  });
}

var detectPerson = function (image) {
  return new Promise((resolve) => {
    var params = {
      SimilarityThreshold: 90, 
      SourceImage: {
        Bytes: null
      }, 
      TargetImage: {
        Bytes: new Buffer.from(image, 'base64'),
      },
    };

    var searchingProcess = []
    var personInPhoto = []

    persons.forEach(function (person) {
      params.SourceImage.Bytes = new Buffer.from(person.image.image, 'base64')
    
      searchingProcess.push(searchingPerson(person.name, params))
    })
    
    Promise.all(searchingProcess).then(function(results) {
    
      results.forEach(function(result) {
        if (result.FaceMatches.length > 0) {
          personInPhoto.push(result.name)
        }
      })

      resolve(personInPhoto)

      return personInPhoto
    });
  });
}

module.exports = detectPerson