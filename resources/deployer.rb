# GeoEngineer Resources For Step Function Deployer
# GEO_ENV=development bundle exec geo apply resources/deployer.rb

########################################
###           ENVIRONMENT            ###
########################################

env = environment('development') {
  region      ENV.fetch('AWS_REGION')
  account_id  ENV.fetch('AWS_ACCOUNT_ID')
}

########################################
###            PROJECT               ###
########################################
project = project('coinbase', 'fenrir') {
  environments 'development'
  tags {
    ProjectName "coinbase/fenrir"
    ConfigName "development"
    DeployWith "step-deployer"
    self[:org] = "coinbase"
    self[:project] = "fenrir"
  }
}

context = {
  assumed_role_name: "coinbase-fenrir-assumed",
  assumable_from: [ ENV['AWS_ACCOUNT_ID'] ],
  assumed_policy_file: "#{__dir__}/assumed_policy.json.erb"
}

project.from_template('bifrost_deployer', 'fenrir', {
  lambda_policy_file: "#{__dir__}/lambda_policy.json.erb",
  lambda_policy_context: context
})

# The assumed role exists in all environments
project.from_template('step_assumed', 'assumed', context)

