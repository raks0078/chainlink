import { ApiResponse } from 'utils/json-api-client'
import { ResourceObject } from 'json-api-normalizer'
import { JobSpecV2 } from 'core/store/models'
import { fluxMonitorJobV2, ocrJobSpecV2 } from './jobSpecV2'

function getRandomInt(max: number) {
  return Math.floor(Math.random() * Math.floor(max))
}

export const jsonApiJobSpecsV2 = (
  jobs: ResourceObject<JobSpecV2>[] = [],
  count?: number,
) => {
  const rc = count || jobs.length

  return {
    data: jobs,
    meta: { count: rc },
  } as ApiResponse<JobSpecV2[]>
}

export const ocrJobResource = (
  job: Partial<
    JobSpecV2['offChainReportingOracleSpec'] & { id?: string; name?: string }
  >,
) => {
  const id = job.id || getRandomInt(1_000_000).toString()

  return {
    type: 'jobs',
    id,
    attributes: {
      ...ocrJobSpecV2(job),
      name: job.name,
    },
  } as ResourceObject<JobSpecV2>
}

export const fluxMonitorJobResource = (
  job: Partial<JobSpecV2['fluxMonitorSpec'] & { id?: string; name?: string }>,
) => {
  const id = job.id || getRandomInt(1_000_000).toString()

  return {
    type: 'jobs',
    id,
    attributes: fluxMonitorJobV2(job, { name: job.name }),
  } as ResourceObject<JobSpecV2>
}
